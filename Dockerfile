ARG GO_VERSION=1.27.3
FROM --platform=$BUILDPLATFORM golang:${GO_VERSION}-alpine AS build
ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT
WORKDIR /build
COPY ./ /build
RUN apk add --update --no-cache git make && \
    CGO_ENABLED=0 GOOS="$TARGETOS" GOARCH="$TARGETARCH" GOARM="${TARGETVARIANT#v}" make build

FROM alpine:latest
COPY --from=build /build/chanify /usr/local/bin/chanify
ENTRYPOINT ["/usr/local/bin/chanify"]
CMD ["serve"]
