FROM golang:1.10-alpine as builder

ARG VERSION

RUN apk add --no-cache git ca-certificates

RUN git clone --branch "v1.1" --single-branch --depth 1 \
    https://github.com/korylprince/fileenv.git /go/src/github.com/korylprince/fileenv

RUN git clone --branch "$VERSION" --single-branch --depth 1 \
    https://github.com/korylprince/chronicle-server.git  /go/src/github.com/korylprince/chronicle-server

RUN go install github.com/korylprince/fileenv
RUN go install github.com/korylprince/chronicle-server

FROM alpine:3.7

RUN apk add --no-cache ca-certificates

COPY --from=builder /go/bin/fileenv /
COPY --from=builder /go/bin/chronicle-server /
COPY setenv.sh /

CMD ["/fileenv", "sh", "/setenv.sh", "/chronicle-server"]
