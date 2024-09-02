# Build Geth in a stock Go builder container
FROM golang:1.22-alpine as builder

RUN apk add --no-cache gcc musl-dev linux-headers git

WORKDIR /bm-cli-tool

ADD . .

RUN cd cmd && go build -o /bm-cli-tool/cli .

FROM alpine:latest

COPY --from=builder /bm-cli-tool/cli /usr/local/bin/

ENTRYPOINT [ "cli" ]