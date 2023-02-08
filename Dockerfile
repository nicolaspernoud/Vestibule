# Dockerfile from https://github.com/chemidy/smallest-secured-golang-docker-image

##################################
# STEP 1 build executable binary #
##################################

FROM golang:1.20-alpine3.17 as builder

# Install git + SSL ca certificates.
# Git is required for fetching the dependencies.
# Ca-certificates is required to call HTTPS endpoints.
RUN apk update && apk add --no-cache git ca-certificates tzdata libcap mailcap && update-ca-certificates

# Create appuser
ENV USER=appuser
ENV UID=1000
# See https://stackoverflow.com/a/55757473/12429735
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "${UID}" \
    "${USER}"

WORKDIR /app

ADD . .

RUN chown -Rf "${UID}" ./*

# Get dependencies and run tests
RUN go version
RUN go get -d -v && CGO_ENABLED=0 go test ./...

# Build the binary
RUN CGO_ENABLED=0 go build \
    -ldflags='-w -s -extldflags "-static"' -a \
    -o /app/vestibule .

# Allow running on ports < 1000
RUN setcap cap_net_bind_service=+ep /app/vestibule

##############################
# STEP 2 build a small image #
##############################
FROM scratch

WORKDIR /app

# Import global resources from builder
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group
COPY --from=builder /etc/mime.types /etc/mime.types

# Copy static executable and application resources
COPY --from=builder /app/vestibule /app/vestibule
COPY --from=builder /app/dev_certificates /app/dev_certificates
COPY --from=builder /app/web /app/web
COPY --from=builder /app/configs /app/configs

# Use an unprivileged user.
USER appuser:appuser

# Run the binary
ENTRYPOINT ["./vestibule"]
