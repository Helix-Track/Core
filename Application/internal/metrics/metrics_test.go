package metrics

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestNewMetrics(t *testing.T) {
	metrics := NewMetrics()
	assert.NotNil(t, metrics)
	assert.NotNil(t, metrics.endpointMetrics)
	assert.False(t, metrics.startTime.IsZero())
}

func TestMetrics_RecordRequest(t *testing.T) {
	metrics := NewMetrics()

	// Record a successful request
	metrics.RecordRequest(100*time.Millisecond, 200, "/api/test", "GET")

	summary := metrics.GetSummary(false)
	assert.Equal(t, int64(1), summary.TotalRequests)
	assert.Equal(t, int64(1), summary.SuccessfulRequests)
	assert.Equal(t, int64(0), summary.FailedRequests)
	assert.Equal(t, int64(1), summary.Status2xx)
}

func TestMetrics_RecordRequest_StatusCodes(t *testing.T) {
	metrics := NewMetrics()

	// Record different status codes
	metrics.RecordRequest(10*time.Millisecond, 200, "/api/test", "GET")
	metrics.RecordRequest(10*time.Millisecond, 201, "/api/test", "POST")
	metrics.RecordRequest(10*time.Millisecond, 301, "/api/redirect", "GET")
	metrics.RecordRequest(10*time.Millisecond, 404, "/api/notfound", "GET")
	metrics.RecordRequest(10*time.Millisecond, 500, "/api/error", "GET")

	summary := metrics.GetSummary(false)
	assert.Equal(t, int64(5), summary.TotalRequests)
	assert.Equal(t, int64(2), summary.SuccessfulRequests)
	assert.Equal(t, int64(2), summary.FailedRequests)
	assert.Equal(t, int64(2), summary.Status2xx)
	assert.Equal(t, int64(1), summary.Status3xx)
	assert.Equal(t, int64(1), summary.Status4xx)
	assert.Equal(t, int64(1), summary.Status5xx)
}

func TestMetrics_Timing(t *testing.T) {
	metrics := NewMetrics()

	// Record requests with different durations
	metrics.RecordRequest(50*time.Millisecond, 200, "/api/fast", "GET")
	metrics.RecordRequest(200*time.Millisecond, 200, "/api/slow", "GET")
	metrics.RecordRequest(100*time.Millisecond, 200, "/api/medium", "GET")

	summary := metrics.GetSummary(false)
	assert.Equal(t, int64(3), summary.TotalRequests)

	// Average should be around 116ms
	assert.Greater(t, summary.AvgRequestTime, 100*time.Millisecond)
	assert.Less(t, summary.AvgRequestTime, 120*time.Millisecond)

	// Min should be 50ms
	assert.Equal(t, 50*time.Millisecond, summary.MinRequestTime)

	// Max should be 200ms
	assert.Equal(t, 200*time.Millisecond, summary.MaxRequestTime)
}

func TestMetrics_EndpointMetrics(t *testing.T) {
	metrics := NewMetrics()

	// Record requests to different endpoints
	metrics.RecordRequest(10*time.Millisecond, 200, "/api/users", "GET")
	metrics.RecordRequest(20*time.Millisecond, 201, "/api/users", "POST")
	metrics.RecordRequest(15*time.Millisecond, 200, "/api/tickets", "GET")

	summary := metrics.GetSummary(true)
	assert.Equal(t, 3, len(summary.Endpoints)) // GET /api/users, POST /api/users, GET /api/tickets

	// Find the /api/users endpoint
	var usersMetric *EndpointMetrics
	for _, endpoint := range summary.Endpoints {
		if endpoint.Path == "/api/users" && endpoint.Method == "GET" {
			usersMetric = endpoint
			break
		}
	}

	assert.NotNil(t, usersMetric)
	assert.Equal(t, int64(1), usersMetric.Count)
	assert.Equal(t, int64(1), usersMetric.Status2xx)
}

func TestMetrics_RequestsPerSecond(t *testing.T) {
	metrics := NewMetrics()

	// Record multiple requests
	for i := 0; i < 100; i++ {
		metrics.RecordRequest(1*time.Millisecond, 200, "/api/test", "GET")
	}

	// Wait a bit to calculate RPS
	time.Sleep(100 * time.Millisecond)

	summary := metrics.GetSummary(false)
	assert.Greater(t, summary.RequestsPerSecond, 0.0)
}

func TestMetrics_Reset(t *testing.T) {
	metrics := NewMetrics()

	// Record some requests
	metrics.RecordRequest(10*time.Millisecond, 200, "/api/test", "GET")
	metrics.RecordRequest(10*time.Millisecond, 404, "/api/notfound", "GET")

	// Verify metrics exist
	summary := metrics.GetSummary(false)
	assert.Equal(t, int64(2), summary.TotalRequests)

	// Reset metrics
	metrics.Reset()

	// Verify metrics are cleared
	summary = metrics.GetSummary(false)
	assert.Equal(t, int64(0), summary.TotalRequests)
	assert.Equal(t, int64(0), summary.SuccessfulRequests)
	assert.Equal(t, int64(0), summary.FailedRequests)
}

func TestMetrics_Uptime(t *testing.T) {
	metrics := NewMetrics()

	// Wait a bit
	time.Sleep(100 * time.Millisecond)

	summary := metrics.GetSummary(false)
	assert.Greater(t, summary.Uptime, 100*time.Millisecond)
}

func TestMetrics_LastRequestTime(t *testing.T) {
	metrics := NewMetrics()

	// Record a request
	metrics.RecordRequest(10*time.Millisecond, 200, "/api/test", "GET")

	summary := metrics.GetSummary(false)
	assert.False(t, summary.LastRequestTime.IsZero())
	assert.WithinDuration(t, time.Now(), summary.LastRequestTime, 1*time.Second)
}

func TestMetricsMiddleware(t *testing.T) {
	metrics := NewMetrics()
	middleware := MetricsMiddleware(metrics)

	router := gin.New()
	router.Use(middleware)
	router.GET("/test", func(c *gin.Context) {
		time.Sleep(10 * time.Millisecond)
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Make request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	// Verify metrics were recorded
	summary := metrics.GetSummary(false)
	assert.Equal(t, int64(1), summary.TotalRequests)
	assert.Equal(t, int64(1), summary.Status2xx)
	assert.Greater(t, summary.AvgRequestTime, 10*time.Millisecond)
}

func TestMetricsMiddleware_MultipleRequests(t *testing.T) {
	metrics := NewMetrics()
	middleware := MetricsMiddleware(metrics)

	router := gin.New()
	router.Use(middleware)
	router.GET("/users", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"users": []string{}})
	})
	router.GET("/tickets", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"tickets": []string{}})
	})

	// Make multiple requests
	for i := 0; i < 5; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/users", nil)
		router.ServeHTTP(w, req)
	}

	for i := 0; i < 3; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/tickets", nil)
		router.ServeHTTP(w, req)
	}

	// Verify metrics
	summary := metrics.GetSummary(true)
	assert.Equal(t, int64(8), summary.TotalRequests)
	assert.Equal(t, 2, len(summary.Endpoints))
}

func TestGetGlobalMetrics(t *testing.T) {
	metrics := GetGlobalMetrics()
	assert.NotNil(t, metrics)
}

func TestHealthCheck_Check(t *testing.T) {
	hc := &HealthCheck{
		Version: "1.0.0",
		DBPing: func() error {
			return nil
		},
	}

	health := hc.Check(false)
	assert.Equal(t, "healthy", health.Status)
	assert.Equal(t, "1.0.0", health.Version)
	assert.Equal(t, "connected", health.Database)
	assert.Greater(t, health.Uptime, time.Duration(0))
	assert.False(t, health.Timestamp.IsZero())
	assert.Nil(t, health.Metrics)
}

func TestHealthCheck_CheckWithMetrics(t *testing.T) {
	// Record some metrics
	metrics := GetGlobalMetrics()
	metrics.Reset()
	metrics.RecordRequest(10*time.Millisecond, 200, "/api/test", "GET")

	hc := &HealthCheck{
		Version: "1.0.0",
		DBPing: func() error {
			return nil
		},
	}

	health := hc.Check(true)
	assert.Equal(t, "healthy", health.Status)
	assert.NotNil(t, health.Metrics)
	assert.Equal(t, int64(1), health.Metrics.TotalRequests)
}

func TestHealthCheck_UnhealthyDatabase(t *testing.T) {
	hc := &HealthCheck{
		Version: "1.0.0",
		DBPing: func() error {
			return assert.AnError
		},
	}

	health := hc.Check(false)
	assert.Equal(t, "unhealthy", health.Status)
	assert.Equal(t, "disconnected", health.Database)
}

func TestMetrics_ConcurrentAccess(t *testing.T) {
	metrics := NewMetrics()
	done := make(chan bool)

	// Concurrent writes
	for i := 0; i < 100; i++ {
		go func(n int) {
			metrics.RecordRequest(10*time.Millisecond, 200, "/api/test", "GET")
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 100; i++ {
		<-done
	}

	// Verify all requests were recorded
	summary := metrics.GetSummary(false)
	assert.Equal(t, int64(100), summary.TotalRequests)
}

// Benchmark tests
func BenchmarkMetrics_RecordRequest(b *testing.B) {
	metrics := NewMetrics()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		metrics.RecordRequest(10*time.Millisecond, 200, "/api/test", "GET")
	}
}

func BenchmarkMetrics_GetSummary(b *testing.B) {
	metrics := NewMetrics()

	// Pre-populate with requests
	for i := 0; i < 1000; i++ {
		metrics.RecordRequest(10*time.Millisecond, 200, "/api/test", "GET")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		metrics.GetSummary(false)
	}
}

func BenchmarkMetrics_GetSummaryWithEndpoints(b *testing.B) {
	metrics := NewMetrics()

	// Pre-populate with requests to different endpoints
	for i := 0; i < 100; i++ {
		metrics.RecordRequest(10*time.Millisecond, 200, "/api/endpoint1", "GET")
		metrics.RecordRequest(10*time.Millisecond, 200, "/api/endpoint2", "POST")
		metrics.RecordRequest(10*time.Millisecond, 200, "/api/endpoint3", "PUT")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		metrics.GetSummary(true)
	}
}
