package monitoring

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// SetupMonitoringRoutes sets up monitoring and health check endpoints
func SetupMonitoringRoutes(router *gin.Engine, metricsCollector *MetricsCollector, healthChecker *HealthChecker) {
	monitoring := router.Group("/monitoring")
	{
		monitoring.GET("/health", healthHandler(healthChecker))
		monitoring.GET("/metrics", metricsHandler(metricsCollector))
		monitoring.GET("/system", systemMetricsHandler(metricsCollector))
	}
}

// healthHandler returns the health status of the application
func healthHandler(healthChecker *HealthChecker) gin.HandlerFunc {
	return func(c *gin.Context) {
		health := healthChecker.CheckHealth(c.Request.Context())
		
		statusCode := http.StatusOK
		if health.Status != "healthy" {
			statusCode = http.StatusServiceUnavailable
		}
		
		c.JSON(statusCode, health)
	}
}

// metricsHandler returns application metrics
func metricsHandler(metricsCollector *MetricsCollector) gin.HandlerFunc {
	return func(c *gin.Context) {
		metrics := metricsCollector.GetMetrics()
		c.JSON(http.StatusOK, gin.H{
			"application_metrics": metrics,
			"timestamp": metrics.StartTime,
		})
	}
}

// systemMetricsHandler returns system-level metrics
func systemMetricsHandler(metricsCollector *MetricsCollector) gin.HandlerFunc {
	return func(c *gin.Context) {
		systemMetrics := metricsCollector.GetSystemMetrics()
		c.JSON(http.StatusOK, gin.H{
			"system_metrics": systemMetrics,
		})
	}
}