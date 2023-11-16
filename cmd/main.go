package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
	"github.com/rs/zerolog/log"
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

func getInt64FromEnvVar(envVarName string) int64 {
	numberAsString := os.Getenv(envVarName)
	numberAsInt, _ := strconv.Atoi(numberAsString)
	return int64(numberAsInt)
}

func getJobSchedulerPort() int64 {
	return getInt64FromEnvVar("JOB_SCHEDULER_PORT")
}

func getHttpPassword() string {
	return os.Getenv("NETWORKPOLICY_CANARY_PASSWORD")
}

func main() {
	zerolog.DurationFieldInteger = true
	logLevel := zerolog.InfoLevel
	if envLogLevel := os.Getenv("LOG_LEVEL"); len(envLogLevel) > 0 {
		if l, err := zerolog.ParseLevel(envLogLevel); err != nil {
			log.Warn().Err(err).Msgf("Unable to parse LOG_LEVEL %s. Using default log level %s", envLogLevel, logLevel.String())
		} else {
			logLevel = l
		}
	}
	zerolog.SetGlobalLevel(logLevel)

	if envPrettyLog := os.Getenv("PRETTY_LOG"); len(envPrettyLog) > 0 {
		if pretty, err := strconv.ParseBool(os.Getenv("PRETTY_LOG")); err != nil {
			log.Warn().Err(err).Msgf("Unable to parse LOG_LEVEL %s. Using default structured logging.", envPrettyLog)
		} else if pretty {
			log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.TimeOnly})
		}
	}

	logger := log.Logger
	logger.Info().Msgf("Starting radix-canary-golang version %s", Version)

	// Configure request logging
	logHandler := hlog.NewHandler(logger)(
		hlog.RequestHandler("request")(
			hlog.RemoteAddrHandler("ip")(
				hlog.AccessHandler(logRequest)(
					hlog.UserAgentHandler("useragent")(http.DefaultServeMux),
				),
			),
		),
	)

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

	port := os.Getenv("LISTENING_PORT")
	logger.Info().Msgf("Starting server on port %v", port)
	if err := http.ListenAndServe(":"+port, logHandler); err != nil {
		logger.Fatal().Err(err).Send()
	}
}

func logRequest(r *http.Request, status, size int, duration time.Duration) {
	hlog.FromRequest(r).Info().Int("status", status).Int("size", size).Dur("took", duration).Send()
}

func getDomains() []string {
	return []string{"google.com", "microsoft.com", "netflix.com", "slack.com", "apple.com"}
}

func getDnsServers() []string {
	return []string{CloudFlareDnsIp1, CloudFlareDnsIp2, GoogleDnsIp1, GoogleDnsIp2}
}

func urlReturns200(ctx context.Context, url string) bool {
	logger := log.Ctx(ctx)
	logger.Debug().Msgf("Sending request to %s", url)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create request")
		return false
	}
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to send request")
		return false
	}
	defer response.Body.Close()
	return response.StatusCode == 200
}

func getGolangCanaryFQDN(ctx context.Context) string {
	logger := log.Ctx(ctx)
	radixDnsZone, isDefined := os.LookupEnv("RADIX_DNS_ZONE")
	if !isDefined {
		logger.Error().Msg("Could not find RADIX_DNS_ZONE")
	}
	clusterName, isDefined := os.LookupEnv("RADIX_CLUSTERNAME")
	if !isDefined {
		logger.Error().Msg("Could not find RADIX_CLUSTERNAME")
	}
	return fmt.Sprintf("https://www-radix-canary-golang-prod.%s.%s", clusterName, radixDnsZone)
}

// testInternalDns iterates over multiple high profile domains. If any of the domains can be resolved, the test passes.
func testInternalDns(w http.ResponseWriter, r *http.Request) {
	logger := hlog.FromRequest(r)
	domains := getDomains()
	for _, domain := range domains {
		logger.Debug().Msgf("Resolving IP for domain %s with default dns server", domain)
		ips, err := net.LookupIP(domain)
		if err == nil && ips != nil {
			logger.Debug().Msgf("Successfully resolved IP for domain %s with default dns server", domain)
			Health(w, r)
			return
		}
		logger.Warn().Err(err).Msgf("Failed to resolve IP for domain %s with default dns server", domain)
	}
	Error(w, r)
}

// testPublicDns iterates over multiple public DNS servers and multiple high profile domains. This test passes
// if any DNS server can resolve any domain. The test will only fail if every DNS server fails on every domain.
func testPublicDns(w http.ResponseWriter, r *http.Request) {
	logger := hlog.FromRequest(r)
	domains := getDomains()
	dnsServers := getDnsServers()
	for _, domain := range domains {
		for _, dnsServer := range dnsServers {
			logger.Debug().Msgf("Resolving IP for domain %s with dns server %s", domain, dnsServer)
			resolver := &net.Resolver{
				PreferGo: true,
				Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
					d := net.Dialer{
						Timeout: time.Millisecond * time.Duration(10000),
					}
					return d.DialContext(ctx, network, fmt.Sprintf("%s:%d", dnsServer, 53))
				},
			}
			ips, err := resolver.LookupHost(r.Context(), domain)
			if err == nil && ips != nil {
				logger.Debug().Msgf("Successfully resolved IP for domain %s with dns server %s", domain, dnsServer)
				Health(w, r)
				return
			}
			logger.Warn().Err(err).Msgf("Failed to resolve IP for domain %s with dns server %s", domain, dnsServer)
		}
	}
	Error(w, r)
}

func testJobScheduler(w http.ResponseWriter, r *http.Request) {
	url := fmt.Sprintf("http://%s:%d%s", JobSchedulerFQDN, getJobSchedulerPort(), JobsPath)
	if urlReturns200(r.Context(), url) {
		Health(w, r)
		return
	}
	Error(w, r)
}

func startJobBatch(w http.ResponseWriter, r *http.Request) {
	logger := hlog.FromRequest(r)
	if requestIsAuthorized(r) {
		url := fmt.Sprintf("http://%s:%d%s", JobSchedulerFQDN, getJobSchedulerPort(), BatchesPath)
		logger.Debug().Msgf("Sending request to %s", url)
		jsonStr := []byte(`{  "jobScheduleDescriptions": [    {      "timeLimitSeconds": 1    }  ]}`)
		response, err := http.Post(url, "application/json", bytes.NewBuffer(jsonStr))
		if err != nil {
			logger.Error().Err(err).Msgf("Failed to send request to %s", url)
			Error(w, r)
			return
		}
		defer response.Body.Close()
		if response.StatusCode == 200 {
			RelayResponse(w, r, response)
			return
		}
		Error(w, r)
	} else {
		logger.Error().Msg("Received unauthorized request")
		Unauthorized(w, r)
	}
}

func testRadixSite(w http.ResponseWriter, r *http.Request) {
	url := getGolangCanaryFQDN(r.Context())
	if urlReturns200(r.Context(), url) {
		Health(w, r)
		return
	}
	Error(w, r)
}

// testExternalWebsite iterates over multiple high profile domains. If any of the domains responds with status code 200, the test passes.
func testExternalWebsite(w http.ResponseWriter, r *http.Request) {
	for _, domain := range getDomains() {
		url := fmt.Sprintf("https://%s", domain)
		if urlReturns200(r.Context(), url) {
			Health(w, r)
			return
		}
	}
	Error(w, r)
}

// Index handler returns a simple front page
func Index(w http.ResponseWriter, r *http.Request) {
	// Increase request count
	requestCount++

	fmt.Fprintf(w, "<h1>Radix Canary App v %s</h1>", Version)
}

// Health handler returns a simple status code indicating system health
func Health(w http.ResponseWriter, r *http.Request) {
	logger := hlog.FromRequest(r)

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

		logger.Error().Err(err).Msg("Unable to encode JSON")

		errorCount++
		return
	}

	// Write JSON to client
	fmt.Fprintf(w, "%s", healthJSON)
}

func requestIsAuthorized(request *http.Request) bool {
	reqToken := request.Header.Get("Authorization")
	splitToken := strings.Split(reqToken, "Bearer ")
	if len(splitToken) < 2 {
		return false
	}
	reqToken = splitToken[1]
	return reqToken == getHttpPassword()
}

func RelayResponse(w http.ResponseWriter, r *http.Request, res *http.Response) {
	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not read response: %s\n", err)

		errorCount++
		return
	}
	fmt.Fprintf(w, "%s", body)
	w.WriteHeader(res.StatusCode)
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
	logger := hlog.FromRequest(r)
	requestCount++
	err := errors.New("can't fulfil request")
	errorJSON, _ := json.Marshal(map[string]interface{}{"Error": fmt.Sprintf("%s", err)})
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintf(w, "%s", errorJSON)
	logger.Error().Err(err).Msg("Server error")
	errorCount++
}

// Unauthorized handler returns an error
func Unauthorized(w http.ResponseWriter, r *http.Request) {
	logger := hlog.FromRequest(r)
	requestCount++
	err := errors.New("Unauthorized request")
	errorJSON, _ := json.Marshal(map[string]interface{}{"Error": fmt.Sprintf("%s", err)})
	w.WriteHeader(http.StatusUnauthorized)
	fmt.Fprintf(w, "%s", errorJSON)
	logger.Error().Err(err).Msg("Server error")
	errorCount++
}

// Echo handler returns the incomming request with headers
func Echo(w http.ResponseWriter, r *http.Request) {
	logger := hlog.FromRequest(r)
	logger.Info().Msgf("%+v", r)

	requestCount++

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

		logger.Error().Err(err).Msg("Unable to encode request JSON")

		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%s", errorJSON)

		errorCount++
		return
	}

	// Write JSON to client
	fmt.Fprintf(w, "%s", requestJSON)
}
