# Radix Canary (Golang)

## Purpose

To have a simple example application that can be used:

- **Primary** As an endpoint for continous monitoring to establish stability and performance baselines and Radix platform SLA measurements
- **Secondary** To verify build and deployment pipelines in Radix
- __Bonus__ Demonstrate some best practices of an application that is to be deployed on the Radix platform

## Inspiration, models, mental frameworks

As with any new application we try to follow the general tenants of the 12FA - [The Twelve-Factor App](https://12factor.net/)

## Running locally

  `go run src/equinor.com/radix-canary-golang/cmd/main.go`

## Building and running using Docker

Build image:

  `docker build --tag radix-canary-golang:latest -f Dockerfile .`

Run image:

  `docker run -p 5000:5000 -e "LISTEN_PORT=5000" radix-canary-golang:latest`

## Features

- Server port configured by environment variables
- Log output to `stdout` and `stderr`
- /health - returns Status: 200
- /metrics - returns number of requests and errors
- /error - increases error count and returns HTTP 500 with an error

The /health endpoint is a common pattern and is used by load balancers and service discovery to determine if a node should receive requests. Read more on [microservices.io](http://microservices.io/patterns/observability/health-check-api.html)

The /metrics endpoint is also a common pattern and is used by scrapers, such as Prometheus, to gather application metrics. PS: Right now this returns JSON, which is NOT the format Prometheus expects.

The /error endpoint is just an example on how to return a different HTTP status code and some payload.

## Further reading

[How I write Go HTTP services after seven years](https://medium.com/statuscode/how-i-write-go-http-services-after-seven-years-37c208122831) is a good source of inspiration for patterns on building larger Golang web services.