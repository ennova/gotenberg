package printer

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/devtool"
	"github.com/mafredri/cdp/protocol/network"
	"github.com/mafredri/cdp/protocol/page"
	"github.com/mafredri/cdp/protocol/runtime"
	"github.com/mafredri/cdp/protocol/target"
	"github.com/mafredri/cdp/rpcc"
	"github.com/thecodingmachine/gotenberg/internal/pkg/conf"
	"github.com/thecodingmachine/gotenberg/internal/pkg/xcontext"
	"github.com/thecodingmachine/gotenberg/internal/pkg/xerror"
	"github.com/thecodingmachine/gotenberg/internal/pkg/xlog"
	"github.com/thecodingmachine/gotenberg/internal/pkg/xtime"
	"golang.org/x/sync/errgroup"
)

type chromePrinter struct {
	logger xlog.Logger
	url    string
	opts   ChromePrinterOptions
}

// ChromePrinterOptions helps customizing the
// Google Chrome Printer behaviour.
type ChromePrinterOptions struct {
	WaitTimeout        float64
	WaitDelay          float64
	WaitJSRenderStatus string
	HeaderHTML         string
	FooterHTML         string
	PaperWidth         float64
	PaperHeight        float64
	MarginTop          float64
	MarginBottom       float64
	MarginLeft         float64
	MarginRight        float64
	Landscape          bool
	PageRanges         string
	RpccBufferSize     int64
	CustomHTTPHeaders  map[string]string
	Scale              float64
	MaxConnections     int64
	WaitForConnection  bool
}

// DefaultChromePrinterOptions returns the default
// Google Chrome Printer options.
func DefaultChromePrinterOptions(config conf.Config) ChromePrinterOptions {
	const defaultHeaderFooterHTML string = "<html><head></head><body></body></html>"
	return ChromePrinterOptions{
		WaitTimeout:        config.DefaultWaitTimeout(),
		WaitDelay:          0.0,
		WaitJSRenderStatus: "",
		HeaderHTML:         defaultHeaderFooterHTML,
		FooterHTML:         defaultHeaderFooterHTML,
		PaperWidth:         8.27,
		PaperHeight:        11.7,
		MarginTop:          1.0,
		MarginBottom:       1.0,
		MarginLeft:         1.0,
		MarginRight:        1.0,
		Landscape:          false,
		PageRanges:         "",
		RpccBufferSize:     config.DefaultGoogleChromeRpccBufferSize(),
		CustomHTTPHeaders:  make(map[string]string),
		Scale:              1.0,
		MaxConnections:     config.GoogleChromeMaxConnections(),
		WaitForConnection:  config.GoogleChromeWaitForConnection(),
	}
}

// nolint: gochecknoglobals
var lockChrome = make(chan struct{}, 1)

// nolint: gochecknoglobals
var devtConnections int64

func (p chromePrinter) Print(destination string) error {
	const op string = "printer.chromePrinter.Print"
	logOptions(p.logger, p.opts)
	ctx, cancel := xcontext.WithTimeout(p.logger, p.opts.WaitTimeout+p.opts.WaitDelay)
	defer cancel()
	resolver := func() error {
		devt, err := devtool.New("http://localhost:9222").Version(ctx)
		if err != nil {
			return err
		}
		// connect to WebSocket URL (page) that speaks the Chrome DevTools Protocol.
		devtConn, err := rpcc.DialContext(ctx, devt.WebSocketDebuggerURL)
		if err != nil {
			return err
		}
		defer devtConn.Close()
		// create a new CDP Client that uses conn.
		devtClient := cdp.NewClient(devtConn)
		createBrowserContextArgs := target.NewCreateBrowserContextArgs()
		newContextTarget, err := devtClient.Target.CreateBrowserContext(ctx, createBrowserContextArgs)
		if err != nil {
			return err
		}
		/*
			close the browser context when done.
			we're not using the "default" context
			as it may timeout before actually closing
			the browser context.
			see: https://github.com/mafredri/cdp/issues/101#issuecomment-524533670
		*/
		disposeBrowserContextArgs := target.NewDisposeBrowserContextArgs(newContextTarget.BrowserContextID)
		defer devtClient.Target.DisposeBrowserContext(context.Background(), disposeBrowserContextArgs) // nolint: errcheck
		// create a new blank target with the new browser context.
		createTargetArgs := target.
			NewCreateTargetArgs("about:blank").
			SetBrowserContextID(newContextTarget.BrowserContextID)
		newTarget, err := devtClient.Target.CreateTarget(ctx, createTargetArgs)
		if err != nil {
			return err
		}
		// connect the client to the new target.
		newTargetWsURL := fmt.Sprintf("ws://127.0.0.1:9222/devtools/page/%s", newTarget.TargetID)
		newContextConn, err := rpcc.DialContext(
			ctx,
			newTargetWsURL,
			/*
				see:
				https://github.com/thecodingmachine/gotenberg/issues/108
				https://github.com/mafredri/cdp/issues/4
				https://github.com/ChromeDevTools/devtools-protocol/issues/24
			*/
			rpcc.WithWriteBufferSize(int(p.opts.RpccBufferSize)),
			rpcc.WithCompression(),
		)
		if err != nil {
			return err
		}
		defer newContextConn.Close()
		// create a new CDP Client that uses newContextConn.
		targetClient := cdp.NewClient(newContextConn)
		/*
			close the target when done.
			we're not using the "default" context
			as it may timeout before actually closing
			the target.
			see: https://github.com/mafredri/cdp/issues/101#issuecomment-524533670
		*/
		closeTargetArgs := target.NewCloseTargetArgs(newTarget.TargetID)
		defer targetClient.Target.CloseTarget(context.Background(), closeTargetArgs) // nolint: errcheck
		// enable all events.
		if err := p.enableEvents(ctx, targetClient); err != nil {
			return err
		}
		// add custom headers (if any).
		if err := p.setCustomHTTPHeaders(ctx, targetClient); err != nil {
			return err
		}
		// listen for crashes
		crashEvent, err := targetClient.Inspector.TargetCrashed(ctx)
		if err != nil {
			return err
		}
		// listen for exceptions
		exceptionEvent, err := targetClient.Runtime.ExceptionThrown(ctx)
		if err != nil {
			return err
		}
		// listen for console messages
		consoleEvent, err := targetClient.Runtime.ConsoleAPICalled(ctx)
		if err != nil {
			return err
		}
		// listen for network requests, responses, and failures
		requestWillBeSentEvent, err := targetClient.Network.RequestWillBeSent(ctx)
		if err != nil {
			return err
		}
		responseReceivedEvent, err := targetClient.Network.ResponseReceived(ctx)
		if err != nil {
			return err
		}
		loadingFailedEvent, err := targetClient.Network.LoadingFailed(ctx)
		if err != nil {
			return err
		}

		var cancelOperation context.CancelFunc

		waiter := func() error {
			// stop listening to async events when we are done waiting
			defer crashEvent.Close()
			defer exceptionEvent.Close()
			defer consoleEvent.Close()
			defer requestWillBeSentEvent.Close()
			defer responseReceivedEvent.Close()
			defer loadingFailedEvent.Close()
			// setup cancel context
			ctx, cancel := context.WithCancel(ctx)
			defer cancel()
			cancelOperation = cancel
			// listen for all events.
			if err := p.listenEvents(ctx, targetClient); err != nil {
				if strings.Contains(err.Error(), "context canceled") {
					return nil
				}
				return err
			}
			// apply a wait delay (if any).
			if p.opts.WaitDelay > 0.0 {
				// wait for a given amount of time (useful for javascript delay).
				p.logger.DebugOpf(op, "applying a wait delay of '%.2fs'...", p.opts.WaitDelay)
				sleep(ctx, xtime.Duration(p.opts.WaitDelay))
			} else {
				p.logger.DebugOp(op, "no wait delay to apply, moving on...")
			}

			if p.opts.WaitJSRenderStatus != "" {
				p.logger.DebugOp(op, "wait for receiving JS render done status"+p.opts.WaitJSRenderStatus)
				if err := Wait(ctx, targetClient, "window.status === '"+p.opts.WaitJSRenderStatus+"'"); err != nil {
					if strings.Contains(err.Error(), "context canceled") {
						return nil
					}
					return err
				}
			}
			return nil
		}

		crashListener := func() error {
			for {
				_, err := crashEvent.Recv()
				if err != nil {
					if strings.Contains(err.Error(), "rpcc: the stream is closing") {
						return nil
					}
					return err
				}
				p.logger.DebugOp(op, "event 'targetCrashed' received")
				cancelOperation()
				return xerror.Invalid(
					op,
					"target has crashed",
					nil,
				)
			}
		}

		exceptionListener := func() error {
			for {
				exception, err := exceptionEvent.Recv()
				if err != nil {
					if strings.Contains(err.Error(), "rpcc: the stream is closing") {
						return nil
					}
					return err
				}
				p.logger.DebugOpf(op, "event 'exceptionThrown' received: %s", exception.ExceptionDetails)
				cancelOperation()
				return xerror.Invalid(
					op,
					exception.ExceptionDetails.Error(),
					nil,
				)
			}
		}

		consoleListener := func() error {
			for {
				log, err := consoleEvent.Recv()
				if err != nil {
					if strings.Contains(err.Error(), "rpcc: the stream is closing") {
						return nil
					}
					return err
				}
				p.logger.DebugOpf(op, "event 'consoleAPICalled' received: %s %s", log.Type, log.Args)
			}
		}

		requestURLs := make(map[network.RequestID]string)
		requestURLsMutex := sync.RWMutex{}

		requestWillBeSentListener := func() error {
			for {
				event, err := requestWillBeSentEvent.Recv()
				if err != nil {
					if strings.Contains(err.Error(), "rpcc: the stream is closing") {
						return nil
					}
					return err
				}
				p.logger.DebugOpf(op, "event 'requestWillBeSent' received: %s %s", event.RequestID, event.Request.URL)
				requestURLsMutex.Lock()
				requestURLs[event.RequestID] = event.Request.URL
				requestURLsMutex.Unlock()
			}
		}

		requestErrorMessages := make(map[network.RequestID]string)
		requestErrorMessagesMutex := sync.RWMutex{}

		responseReceivedListener := func() error {
			for {
				event, err := responseReceivedEvent.Recv()
				if err != nil {
					if strings.Contains(err.Error(), "rpcc: the stream is closing") {
						return nil
					}
					return err
				}

				requestURLsMutex.RLock()
				url := requestURLs[event.RequestID]
				requestURLsMutex.RUnlock()
				msg := fmt.Sprintf("%d %s", event.Response.Status, event.Response.StatusText)
				p.logger.DebugOpf(op, "event 'responseReceived' received: %s: %s", url, msg)

				if event.Response.Status < 400 {
					continue
				}

				requestErrorMessagesMutex.Lock()
				if value, ok := requestErrorMessages[event.RequestID]; !ok || value == "net::ERR_ABORTED" {
					requestErrorMessages[event.RequestID] = msg
					cancelOperation()
				}
				requestErrorMessagesMutex.Unlock()
			}
		}

		loadingFailedListener := func() error {
			for {
				event, err := loadingFailedEvent.Recv()
				if err != nil {
					if strings.Contains(err.Error(), "rpcc: the stream is closing") {
						return nil
					}
					return err
				}

				requestURLsMutex.RLock()
				url := requestURLs[event.RequestID]
				requestURLsMutex.RUnlock()
				msg := fmt.Sprintf("%s", event.ErrorText)
				p.logger.DebugOpf(op, "event 'loadingFailed' received: %s: %s", url, msg)

				requestErrorMessagesMutex.Lock()
				if _, ok := requestErrorMessages[event.RequestID]; !ok {
					requestErrorMessages[event.RequestID] = msg
					cancelOperation()
				}
				requestErrorMessagesMutex.Unlock()
			}
		}

		if err := runBatch(
			ctx,
			crashListener,
			exceptionListener,
			consoleListener,
			requestWillBeSentListener,
			responseReceivedListener,
			loadingFailedListener,
			waiter,
		); err != nil {
			return err
		}

		if len(requestErrorMessages) > 0 {
			msg := ""
			for requestID, message := range requestErrorMessages {
				url := requestURLs[requestID]
				if len(msg) > 0 {
					msg += "\n"
				}
				msg += fmt.Sprintf("%s: %s", url, message)
			}
			return xerror.Invalid(
				op,
				msg,
				nil,
			)
		}

		// listen for crashes
		crashEvent, err = targetClient.Inspector.TargetCrashed(ctx)
		if err != nil {
			return err
		}

		printer := func() error {
			// stop listening to crashes when we are done printing
			defer crashEvent.Close()

			// setup cancel context
			ctx, cancel := context.WithCancel(ctx)
			defer cancel()
			cancelOperation = cancel

			printToPdfArgs := page.NewPrintToPDFArgs().
				SetTransferMode("ReturnAsStream").
				SetPaperWidth(p.opts.PaperWidth).
				SetPaperHeight(p.opts.PaperHeight).
				SetMarginTop(p.opts.MarginTop).
				SetMarginBottom(p.opts.MarginBottom).
				SetMarginLeft(p.opts.MarginLeft).
				SetMarginRight(p.opts.MarginRight).
				SetLandscape(p.opts.Landscape).
				SetDisplayHeaderFooter(true).
				SetHeaderTemplate(p.opts.HeaderHTML).
				SetFooterTemplate(p.opts.FooterHTML).
				SetPrintBackground(true).
				SetScale(p.opts.Scale)
			if p.opts.PageRanges != "" {
				printToPdfArgs.SetPageRanges(p.opts.PageRanges)
			}
			// printToPDF the page to PDF.
			p.logger.DebugOp(op, "starting PrintToPDF")
			printToPDF, err := targetClient.Page.PrintToPDF(
				ctx,
				printToPdfArgs,
			)
			if err != nil {
				// find a way to check it in the handlers?
				if strings.Contains(err.Error(), "Page range syntax error") {
					return xerror.Invalid(
						op,
						fmt.Sprintf("'%s' is not a valid Google Chrome page ranges", p.opts.PageRanges),
						err,
					)
				}
				if strings.Contains(err.Error(), "rpcc: message too large") {
					return xerror.Invalid(
						op,
						fmt.Sprintf(
							"'%d' bytes are not enough: increase the Google Chrome rpcc buffer size (up to 100 MB)",
							p.opts.RpccBufferSize,
						),
						err,
					)
				}
				return err
			}

			p.logger.DebugOp(op, "streaming PDF from Chrome")
			streamReader := targetClient.NewIOStreamReader(ctx, *printToPDF.Stream)
			reader := bufio.NewReader(streamReader)
			file, err := os.OpenFile(destination, os.O_CREATE|os.O_WRONLY, 0600)
			if err != nil {
				return err
			}
			if _, err = reader.WriteTo(file); err != nil {
				return err
			}
			if err = file.Close(); err != nil {
				return err
			}
			p.logger.DebugOp(op, "streaming complete")

			return nil
		}

		if err := runBatch(
			ctx,
			crashListener,
			printer,
		); err != nil {
			return err
		}

		return nil
	}
	if devtConnections+1 < p.opts.MaxConnections {
		p.logger.DebugOp(op, "skipping lock acquisition...")
		devtConnections++
		err := resolver()
		devtConnections--
		if err != nil {
			return xcontext.MustHandleError(
				ctx,
				xerror.New(op, err),
			)
		}
		return nil
	}
	if devtConnections >= p.opts.MaxConnections && !p.opts.WaitForConnection {
		return xerror.Invalid(
			op,
			"no available connections",
			nil,
		)
	}
	p.logger.DebugOp(op, "waiting lock to be acquired...")
	select {
	case lockChrome <- struct{}{}:
		// lock acquired.
		p.logger.DebugOp(op, "lock acquired")
		devtConnections++
		err := resolver()
		devtConnections--
		<-lockChrome // we release the lock.
		if err != nil {
			return xcontext.MustHandleError(
				ctx,
				xerror.New(op, err),
			)
		}
		return nil
	case <-ctx.Done():
		// failed to acquire lock before
		// deadline.
		p.logger.DebugOp(op, "failed to acquire lock before context.Context deadline")
		return xcontext.MustHandleError(
			ctx,
			ctx.Err(),
		)
	}
}

func (p chromePrinter) enableEvents(ctx context.Context, client *cdp.Client) error {
	const op string = "printer.chromePrinter.enableEvents"
	// enable all the domain events that we're interested in.
	if err := runBatch(
		ctx,
		func() error { return client.DOM.Enable(ctx) },
		func() error { return client.Network.Enable(ctx, network.NewEnableArgs()) },
		func() error { return client.Page.Enable(ctx) },
		func() error {
			return client.Page.SetLifecycleEventsEnabled(ctx, page.NewSetLifecycleEventsEnabledArgs(true))
		},
		func() error { return client.Runtime.Enable(ctx) },
	); err != nil {
		return xerror.New(op, err)
	}
	return nil
}

func (p chromePrinter) setCustomHTTPHeaders(ctx context.Context, client *cdp.Client) error {
	const op string = "printer.chromePrinter.setCustomHTTPHeaders"
	resolver := func() error {
		if len(p.opts.CustomHTTPHeaders) == 0 {
			p.logger.DebugOp(op, "skipping custom HTTP headers as none have been provided...")
			return nil
		}
		customHTTPHeaders := make(map[string]string)
		// useless but for the logs.
		for key, value := range p.opts.CustomHTTPHeaders {
			customHTTPHeaders[key] = value
			p.logger.DebugOpf(op, "set '%s' to custom HTTP header '%s'", value, key)
		}
		b, err := json.Marshal(customHTTPHeaders)
		if err != nil {
			return err
		}
		// should always be called after client.Network.Enable.
		return client.Network.SetExtraHTTPHeaders(ctx, network.NewSetExtraHTTPHeadersArgs(b))
	}
	if err := resolver(); err != nil {
		return xerror.New(op, err)
	}
	return nil
}

func (p chromePrinter) listenEvents(ctx context.Context, client *cdp.Client) error {
	const op string = "printer.chromePrinter.listenEvents"
	resolver := func() error {
		// make sure Page events are enabled.
		if err := client.Page.Enable(ctx); err != nil {
			return err
		}
		// make sure Network events are enabled.
		if err := client.Network.Enable(ctx, nil); err != nil {
			return err
		}
		// create all clients for events.
		domContentEventFired, err := client.Page.DOMContentEventFired(ctx)
		if err != nil {
			return err
		}
		defer domContentEventFired.Close()
		loadEventFired, err := client.Page.LoadEventFired(ctx)
		if err != nil {
			return err
		}
		defer loadEventFired.Close()
		lifecycleEvent, err := client.Page.LifecycleEvent(ctx)
		if err != nil {
			return err
		}
		defer lifecycleEvent.Close()
		loadingFinished, err := client.Network.LoadingFinished(ctx)
		if err != nil {
			return err
		}
		defer loadingFinished.Close()
		if _, err := client.Page.Navigate(ctx, page.NewNavigateArgs(p.url)); err != nil {
			return err
		}
		// wait for all events.
		return runBatch(
			ctx,
			func() error {
				_, err := domContentEventFired.Recv()
				if err != nil {
					return err
				}
				p.logger.DebugOp(op, "event 'domContentEventFired' received")
				return nil
			},
			func() error {
				_, err := loadEventFired.Recv()
				if err != nil {
					return err
				}
				p.logger.DebugOp(op, "event 'loadEventFired' received")
				return nil
			},
			func() error {
				const networkIdleEventName string = "networkIdle"
				for {
					ev, err := lifecycleEvent.Recv()
					if err != nil {
						return err
					}
					p.logger.DebugOpf(op, "event '%s' received", ev.Name)
					if ev.Name == networkIdleEventName {
						break
					}
				}
				return nil
			},
			func() error {
				_, err := loadingFinished.Recv()
				if err != nil {
					return err
				}
				p.logger.DebugOp(op, "event 'loadingFinished' received")
				return nil
			},
		)
	}
	if err := resolver(); err != nil {
		return xerror.New(op, err)
	}
	return nil
}

func runBatch(ctx context.Context, fn ...func() error) error {
	// run all functions simultaneously and wait until
	// execution has completed or an error is encountered.
	eg, ctx := errgroup.WithContext(ctx)
	for _, f := range fn {
		eg.Go(f)
	}
	return eg.Wait()
}

// Compile-time checks to ensure type implements desired interfaces.
var (
	_ = Printer(new(chromePrinter))
)

func Eval(ctx context.Context, c *cdp.Client, expr string, out interface{}) error {
	args := runtime.NewEvaluateArgs(expr).
		SetReturnByValue(out != nil)
	return eval(ctx, c, args, out)
}

func EvalPromise(ctx context.Context, c *cdp.Client, expr string, out interface{}) error {
	args := runtime.NewEvaluateArgs(expr).
		SetReturnByValue(out != nil).
		SetAwaitPromise(true)
	return eval(ctx, c, args, out)
}

func eval(ctx context.Context, c *cdp.Client, args *runtime.EvaluateArgs, out interface{}) error {
	reply, err := c.Runtime.Evaluate(ctx, args)
	if err != nil {
		return err
	}
	if reply.ExceptionDetails != nil {
		return reply.ExceptionDetails
	}
	if out == nil {
		return nil
	}
	return json.Unmarshal(reply.Result.Value, out)
}

func Wait(ctx context.Context, c *cdp.Client, expr string) error {
	return poll(ctx, func() (bool, error) {
		var ok bool
		if err := Eval(ctx, c, expr, &ok); err != nil {
			return false, err
		}
		return ok, nil
	})
}

func poll(ctx context.Context, fn func() (bool, error)) error {
	t := time.NewTimer(1 * time.Second)
	if !t.Stop() {
		<-t.C
	}

	for {
		ok, err := fn()
		if err != nil {
			return err
		}
		if ok {
			return nil
		}

		t.Reset(100 * time.Millisecond)
		select {
		case <-t.C:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func sleep(ctx context.Context, delay time.Duration) {
	select {
	case <-ctx.Done():
	case <-time.After(delay):
	}
}
