# Building...

FROM golang:alpine as server-builder

WORKDIR /vestibule

RUN apk update && apk upgrade && \
    apk add --no-cache bash git openssh build-base
ADD . .
RUN go get -d -v && \
    go test ./... && \
    go build

# Running...

FROM alpine

WORKDIR /app

RUN apk update && apk add ca-certificates libcap

COPY --from=server-builder /vestibule/vestibule /app
COPY --from=server-builder /vestibule/dev_certificates /app/dev_certificates
COPY --from=server-builder /vestibule/web /app/web
COPY --from=server-builder /vestibule/configs /app/configs

RUN setcap cap_net_bind_service=+ep vestibule

ENTRYPOINT [ "./vestibule"]
