FROM golang:alpine
WORKDIR /build
COPY ./ /build
RUN apk add --update --no-cache git make && make build

FROM alpine:latest
COPY --from=0 /build/chanify /usr/local/bin/chanify
ENTRYPOINT ["/usr/local/bin/chanify"]
CMD ["serve"]
