package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/handlers"
)

const (
	// HealthStatusOK is a status code for when everything is OK
	HealthStatusOK   = 200
	CloudFlareDnsIp1 = "1.1.1.1"
	CloudFlareDnsIp2 = "1.0.0.1"
	GoogleDnsIp1     = "8.8.8.8"
	GoogleDnsIp2     = "8.8.4.4"
	JobsPath         = "/api/v1/jobs"
    BatchesPath      = "/api/v1/batches"
	JobSchedulerFQDN = "myjob"
)

const (
	// Version is the version number of Radix Canary Golang
	Version = "0.1.18"
	// ListenPort Default port for server to listen on unless specified in environment variable
	ListenPort = "5000"
)

// HealthStatus defines various fields we might include in our health status
type HealthStatus struct {
	Status int
}

// Simple counters for application metrics
var requestCount int64
var errorCount int64

func main() {
	fmt.Printf("Starting radix-canary-golang version %s\n", Version)

	// Retrieving the default nameserver IPs from /etc/resolv.conf

	// Register handler functions to URL paths
	http.HandleFunc("/", Index)
	http.HandleFunc("/health", Health)
	http.HandleFunc("/metrics", Metrics)
	http.HandleFunc("/error", Error)
	http.HandleFunc("/echo", Echo)
	http.HandleFunc("/testpublicdns", testPublicDns)
	http.HandleFunc("/testinternaldns", testInternalDns)
	http.HandleFunc("/testjobscheduler", testJobScheduler)
	http.HandleFunc("/startjobbatch", startJobBatch)
	http.HandleFunc("/testexternalwebsite", testExternalWebsite)
	http.HandleFunc("/testradixsite", testRadixSite)

	// See if listen_port environment variable is set
	port := os.Getenv("LISTEN_PORT")

	// Default port if none given
	if port == "" {
		port = ListenPort
	}

	fmt.Printf("Starting server on port %v\n", port)

	// Start server. Exit fatally on error
	log.Fatal(http.ListenAndServe(":"+port, handlers.CompressHandler(handlers.LoggingHandler(os.Stdout, http.DefaultServeMux))))
}

func getDomains() []string {
	return []string{"google.com", "microsoft.com", "netflix.com", "slack.com", "apple.com"}
}

func getDnsServers() []string {
	return []string{CloudFlareDnsIp1, CloudFlareDnsIp2, GoogleDnsIp1, GoogleDnsIp2}
}

func urlReturns200(url string) bool {
	response, err := http.Get(url)
	if err == nil && response.StatusCode == 200 {
		return true
	}
	return false
}

func getGolangCanaryFQDN() string {
	radixDnsZone, isDefined := os.LookupEnv("RADIX_DNS_ZONE")
	if !isDefined {
		fmt.Println("Could not find RADIX_DNS_ZONE")
	}
	return fmt.Sprintf("https://www-radix-canary-golang-prod.%s", radixDnsZone)
}

// testInternalDns iterates over multiple high profile domains. If any of the domains can be resolved, the test passes.
func testInternalDns(writer http.ResponseWriter, request *http.Request) {
	domains := getDomains()
	for _, domain := range domains {
		ips, err := net.LookupIP(domain)
		if err == nil && ips != nil {
			Health(writer, request)
			return
		}
	}
	Error(writer, request)
}

// testPublicDns iterates over multiple public DNS servers and multiple high profile domains. This test passes
// if any DNS server can resolve any domain. The test will only fail if every DNS server fails on every domain.
func testPublicDns(writer http.ResponseWriter, request *http.Request) {
	domains := getDomains()
	dnsServers := getDnsServers()
	for _, domain := range domains {
		for _, dnsServer := range dnsServers {
			r := &net.Resolver{
				PreferGo: true,
				Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
					d := net.Dialer{
						Timeout: time.Millisecond * time.Duration(10000),
					}
					return d.DialContext(ctx, network, fmt.Sprintf("%s:%d", dnsServer, 53))
				},
			}
			ips, err := r.LookupHost(context.Background(), domain)
			if err == nil && ips != nil {
				Health(writer, request)
				return
			}
		}
	}
	Error(writer, request)
}

func testJobScheduler(writer http.ResponseWriter, request *http.Request) {
	url := fmt.Sprintf("https://%s%s", JobSchedulerFQDN, JobsPath)
	if urlReturns200(url) {
		Health(writer, request)
		return
	}
	Error(writer, request)
}

func startJobBatch(writer http.ResponseWriter, request *http.Request) {
	// curl -X POST "http://127.0.0.1:9000/api/v1/batches" -H  "accept: application/json" -H  "Content-Type: application/json" -d "{  \"jobScheduleDescriptions\": [    {      \"timeLimitSeconds\": 1    }  ]}"
	url := fmt.Sprintf("https://%s%s", JobSchedulerFQDN, BatchesPath)
	jsonStr := []byte(`{  \"jobScheduleDescriptions\": [    {      \"timeLimitSeconds\": 1    }  ]}`)
	
	response, err := http.Post(url, "application/json", bytes.NewBuffer(jsonStr))
	if err == nil && response.StatusCode == 200 {
		Health(writer, request)
		return
	}
	Error(writer, request)
}

func testRadixSite(writer http.ResponseWriter, request *http.Request) {
	url := getGolangCanaryFQDN()
	if urlReturns200(url) {
		Health(writer, request)
		return
	}
	Error(writer, request)
}

func testExternalWebsite(writer http.ResponseWriter, request *http.Request) {
	for _, d := range getDomains() {
		if urlReturns200(fmt.Sprintf("https://%s", d)) {
			Health(writer, request)
			return
		}
	}
	Error(writer, request)
}

func isPortOpen(host string, port string, timeoutSeconds int64) bool {
	timeout := time.Duration(timeoutSeconds * 1000000000)
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, port), timeout)
	defer conn.Close()
	if err != nil {
		return false
	}
	if conn == nil {
		return false
	}
	return true
}

// Index handler returns a simple front page
func Index(w http.ResponseWriter, r *http.Request) {
	// Increase request count
	requestCount++

	fmt.Fprintf(w, "<h1>Radix Canary App v %s</h1>", Version)
}

// Health handler returns a simple status code indicating system health
func Health(w http.ResponseWriter, r *http.Request) {
	// Increase request count
	requestCount++

	// Create a health type instance
	health := HealthStatus{Status: HealthStatusOK}

	// Convert health instance to JSON
	healthJSON, err := json.Marshal(health)

	// Check for errors from JSON conversion
	if err != nil {
		errorJSON, err := json.Marshal(map[string]interface{}{"Error": err})

		// Return error and HTTP status code to client
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%s", errorJSON)

		// Print error to server standard error
		fmt.Fprintf(os.Stderr, "Could not encode JSON: %s\n", err)

		errorCount++
		return
	}

	// Write JSON to client
	fmt.Fprintf(w, "%s", healthJSON)
}

// Metrics handler returns some application metrics in JSON format
func Metrics(w http.ResponseWriter, r *http.Request) {
	requestCount++

	hostname, _ := os.Hostname()

	// Valid label names: [a-zA-Z_][a-zA-Z0-9_]*
	// https://prometheus.io/docs/concepts/data_model/#metric-names-and-labels
	labels := map[string]interface{}{
		"host":      hostname,
		"pid":       os.Getpid(),
		"component": "radix-canary-go",
		"version":   Version,
	}

	var labelsStr string

	for labelName, labelValue := range labels {
		labelsStr += fmt.Sprintf(`%s="%v",`, labelName, labelValue)
	}
	labelsStr = strings.Trim(labelsStr, ",")

	appMetrics := map[string]interface{}{
		"requests_total": requestCount,
		"errors_total":   errorCount,
	}

	for metric, value := range appMetrics {
		fmt.Fprintf(w, "%s{%s} %v\n", metric, labelsStr, value)
	}

}

// Error handler returns an error
func Error(w http.ResponseWriter, r *http.Request) {
	requestCount++

	err := errors.New("Can't fulfil request")

	if err != nil {
		errorJSON, _ := json.Marshal(map[string]interface{}{"Error": fmt.Sprintf("%s", err)})

		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%s", errorJSON)

		fmt.Fprintf(os.Stderr, "Server error: %s\n", err)

		errorCount++
		return
	}
}

// Echo handler returns the incomming request with headers
func Echo(w http.ResponseWriter, r *http.Request) {
	requestCount++

	fmt.Printf("%+v", r)

	request := map[string]interface{}{
		"headers":    r.Header,
		"method":     r.Method,
		"url":        r.URL,
		"requesturi": r.RequestURI,
		"remoteaddr": r.RemoteAddr,
		"body":       r.Body,
	}

	requestJSON, err := json.Marshal(request)

	if err != nil {
		errorJSON, _ := json.Marshal(map[string]interface{}{"Error": err})

		fmt.Fprintf(os.Stderr, "Could not encode request JSON: %v\n", err)

		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%s", errorJSON)

		errorCount++
		return
	}

	// Write JSON to client
	fmt.Fprintf(w, "%s", requestJSON)
}
