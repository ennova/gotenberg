package xhttp

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/thecodingmachine/gotenberg/internal/pkg/conf"
)

// New returns a custom echo.Echo.
func New(config conf.Config) *echo.Echo {
	srv := echo.New()
	srv.HideBanner = true
	srv.HidePort = true

	if config.EnableAuthentication() {
		srv.Use(middleware.BasicAuth(func(username, password string, c echo.Context) (bool, error) {
			if username == config.AuthenticationUsername() && password == config.AuthenticationPassword() {
				return true, nil
			}
			return false, nil
		}))
	}

	srv.Use(contextMiddleware(config))
	srv.Use(loggerMiddleware(config))
	srv.Use(cleanupMiddleware())
	srv.Use(errorMiddleware())
	srv.GET(pingEndpoint(config), pingHandler)
	// srv.POST(mergeEndpoint(config), mergeHandler)
	if config.DisableGoogleChrome() && config.DisableUnoconv() {
		return srv
	}
	if !config.DisableGoogleChrome() {
		srv.POST(htmlEndpoint(config), htmlHandler)
		srv.POST(urlEndpoint(config), urlHandler)
		srv.POST(markdownEndpoint(config), markdownHandler)
	}
	if !config.DisableUnoconv() {
		// srv.POST(officeEndpoint(config), officeHandler)
	}
	return srv
}
