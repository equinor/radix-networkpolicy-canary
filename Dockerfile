FROM golang:1.21-alpine3.18 as builder

WORKDIR /go/src/github.com/equinor/radix-networkpolicy-canary/

# get dependencies
COPY go.mod go.sum ./
RUN go mod download

# copy api code
COPY . .

#Build
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-s -w" -a -installsuffix cgo -o /radix-networkpolicy-canary cmd/main.go

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /radix-networkpolicy-canary /app/radix-networkpolicy-canary

ENV LISTENING_PORT "5000"
EXPOSE 5000
USER 2000
ENTRYPOINT ["/app/radix-networkpolicy-canary"]
