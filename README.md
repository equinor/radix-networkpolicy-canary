# Radix Canary (Golang)

## Purpose

To have a simple example application that can be used:

- **Primary** As an endpoint for continous monitoring to establish stability and performance baselines and Radix platform SLA measurements
- **Secondary** To verify build and deployment pipelines in Radix
- **Secondary** To verify that HTTP requests sent through Kubernetes ingresses and proxies arrive as expected
- __Bonus__ Demonstrate some best practices of an application that is to be deployed on the Radix platform
- __Bonus__ Generate CPU and memory load on the Radix platform


## Inspiration, models, mental frameworks

As with any new application we try to follow the general tenants of the 12FA - [The Twelve-Factor App](https://12factor.net/)

## Running locally

  `go run src/equinor.com/radix-canary-golang/cmd/main.go`

## Building and running using Docker

Build image:

  `docker build --tag radix-canary-golang:latest -f Dockerfile .`

Run image:

  `docker run -p 5000:5000 -e "LISTEN_PORT=5000" radix-canary-golang:latest`

## Deploy to Kubernetes

  `kubectl create -f deployment-service.yaml`

## Register on Radix Platform

  `kubectl create -f radixregistration.yaml`

## Features

- Server port configured by environment variables
- Log output to `stdout` and `stderr`
- /health - returns Status: 200
- /metrics - returns number of requests and errors
- /error - increases error count and returns HTTP 500 with an error
- /echo - returns the incomming request data including headers
- /calculatehashesbcrypt - CPU intensive task that generate and compare Bcrypt hashes
- /calculatehashesscrypt - CPU and memory intensive task that generates Scrypt derived keys 

The /health endpoint is a common pattern and is used by load balancers and service discovery to determine if a node should receive requests. Read more on [microservices.io](http://microservices.io/patterns/observability/health-check-api.html)

The /metrics endpoint is also a common pattern and is used by scrapers, such as Prometheus, to gather application metrics. This application spits out a very basic Prometheus-compatible counter. For production you should use the official Prometheus client libraries to export metrics: https://godoc.org/github.com/prometheus/client_golang/prometheus . Also look for the Radix Monitoring Manual on general information on monitoring and instrumentation. 

The /error endpoint is just an example on how to return a different HTTP status code and some payload.

The /echo endpoint returns the incomming request as seen from the server. This is useful since there might be intermediate proxies that modifies a request before arriving at a server. Using this we can verify that a request arrives as expected.

The /calculatehashesbcrypt and /calculatehashesscrypt endpoints emulate CPU and memory intensive operations. Using these we can try to break things and discover failure modes and practice operations without needing a huge load generator as a client.

## Load testing

k6 run -< k6_test_canary.js

## Further reading

[How I write Go HTTP services after seven years](https://medium.com/statuscode/how-i-write-go-http-services-after-seven-years-37c208122831) is a good source of inspiration for patterns on building larger Golang web services..
