# |--------------------------------------------------------------------------
# | Binary
# |--------------------------------------------------------------------------
# |
# | Builds Gotenberg binary.
# |

ARG DOCKER_REGISTRY
ARG DOCKER_PLATFORM_SUFFIX

FROM ${DOCKER_REGISTRY}/gotenberg:workspace${DOCKER_PLATFORM_SUFFIX} AS workspace

ENV GOOS=linux \
    GOARCH=amd64 \
    CGO_ENABLED=0

# Define our workding outside of $GOPATH (we're using go modules).
WORKDIR /gotenberg/package

# Install module dependencies.
COPY go.mod go.sum ./

RUN --mount=type=cache,id=gotenberg-go-mod,target=/gotenberg/go/pkg/mod \
    go mod download &&\
    go mod verify

# Copy our source code.
COPY internal ./internal
COPY cmd ./cmd

# Build our binary.
ARG VERSION

RUN --mount=type=cache,id=gotenberg-go-mod,target=/gotenberg/go/pkg/mod \
    --mount=type=cache,id=gotenberg-go-build,target=/root/.cache/go-build \
    go build -o gotenberg -ldflags "-X main.version=${VERSION}" cmd/gotenberg/main.go

FROM ${DOCKER_REGISTRY}/gotenberg:base${DOCKER_PLATFORM_SUFFIX}

ARG VERSION

LABEL authors="Julien Neuhart <j.neuhart@thecodingmachine.com>"
LABEL org.opencontainers.image.version="${VERSION}"

# |--------------------------------------------------------------------------
# | Tini
# |--------------------------------------------------------------------------
# |
# | An helper for reaping zombie processes.
# |

ARG TINI_VERSION

ADD https://github.com/krallin/tini/releases/download/v${TINI_VERSION}/tini-static /tini
RUN chmod +x /tini
ENTRYPOINT [ "/tini", "--" ]

# |--------------------------------------------------------------------------
# | Final touch
# |--------------------------------------------------------------------------
# |
# | Last instructions of this build.
# |

COPY --from=workspace /gotenberg/package/gotenberg /usr/local/bin/

USER gotenberg
WORKDIR /gotenberg

EXPOSE 3000
CMD [ "gotenberg" ]
