FROM golang:alpine
WORKDIR /build
COPY ./ /build
RUN apk add --update --no-cache git make && GOOS=linux GOARCH=amd64 make build

FROM alpine:latest
COPY --from=0 /build/chanify /usr/local/bin/chanify
ENTRYPOINT ["/usr/local/bin/chanify"]
CMD ["serve"]
