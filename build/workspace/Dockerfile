ARG DOCKER_REGISTRY
ARG DOCKER_PLATFORM_SUFFIX
ARG GOLANG_VERSION

FROM golang:${GOLANG_VERSION}-bullseye as golang

FROM ${DOCKER_REGISTRY}/gotenberg:base${DOCKER_PLATFORM_SUFFIX}

# |--------------------------------------------------------------------------
# | Common libraries
# |--------------------------------------------------------------------------
# |
# | Libraries used in the build process of this image.
# |

RUN apt-get update && apt-get install -y --no-install-recommends \
        git \
		g++ \
		gcc \
		libc6-dev \
		make \
		pkg-config \
	&& rm -rf /var/lib/apt/lists/*

# |--------------------------------------------------------------------------
# | Golang
# |--------------------------------------------------------------------------
# |
# | Installs Golang.
# |

COPY --from=golang /usr/local/go /usr/local/go

ENV GOPATH /gotenberg/go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" &&\
    chmod -R 777 "$GOPATH"

# |--------------------------------------------------------------------------
# | Final touch
# |--------------------------------------------------------------------------
# |
# | Last instructions of this build.
# |

# Make sure the Gotenber user is able to
# call the Go binary.
USER gotenberg

RUN go version &&\
    go env

USER root