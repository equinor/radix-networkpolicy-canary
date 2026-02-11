# Build stage
FROM --platform=$BUILDPLATFORM docker.io/golang:1.25.7-alpine3.23 AS builder
ARG TARGETARCH
ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=${TARGETARCH}

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download && go mod tidy
COPY . .
RUN go build -ldflags="-s -w" -o /build/radix-networkpolicy-canary ./cmd/main.go

FROM scratch

# Final stage, ref https://github.com/GoogleContainerTools/distroless/blob/main/base/README.md for distroless
FROM gcr.io/distroless/static
ENV LISTENING_PORT="5000"
WORKDIR /app
COPY --from=builder /build/radix-networkpolicy-canary .
EXPOSE 5000
USER 2000
ENTRYPOINT ["/app/radix-networkpolicy-canary"]
