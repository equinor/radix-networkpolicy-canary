package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

const (
	// HealthStatusOK is a status code for when everything is OK
	HealthStatusOK = 200
)

const (
	// Version is the version number of Radix Canary Golang
	Version = "0.1.6"
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
	fmt.Printf("Starting radix-canary-golang\n")

	// Register handler functions to URL paths
	http.HandleFunc("/", Index)
	http.HandleFunc("/health", Health)
	http.HandleFunc("/metrics", Metrics)
	http.HandleFunc("/error", Error)
	http.HandleFunc("/echo", Echo)

	// See if listen_port environment variable is set
	port := os.Getenv("LISTEN_PORT")

	// Default port if none given
	if port == "" {
		port = ListenPort
	}

	fmt.Printf("Starting server on port %v\n", port)

	// Start server. Exit fatally on error
	log.Fatal(http.ListenAndServe(":"+port, nil))
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
