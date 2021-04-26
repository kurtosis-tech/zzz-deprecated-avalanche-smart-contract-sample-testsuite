# We use Debian (rather than Alpine) here because there are certain compatibility issues with Alpine:
# 1) https://github.com/golang/go/issues/28065 that prevents 'go test' from running inside an Alpine container unless you have CGO_ENABLED=0
# 2) Geth needs CGO_ENABLED=1
FROM golang:1.15-buster AS builder

WORKDIR /build
# Copy and download dependencies using go mod
COPY go.mod .
COPY go.sum .
RUN go mod download

# Copy the code into the container
COPY . .

RUN go test ./...

# Build the application
RUN GOOS=linux go build -o testsuite.bin testsuite/main.go

# ============= Execution Stage ================
# We use buster here because we used buster to build, and there are some dynamically-linked lib requirements
FROM debian:buster AS execution

WORKDIR /run

# Copy the code into the container
COPY --from=builder /build/testsuite.bin .

# TODO Switch to exec command form, wrapping arguments with double-quote
CMD ./testsuite.bin \
    --custom-params-json="${CUSTOM_PARAMS_JSON}" \
    --kurtosis-api-socket="${KURTOSIS_API_SOCKET}" \
    --log-level="${LOG_LEVEL}"
