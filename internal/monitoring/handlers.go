package monitoring

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// SetupMonitoringRoutes sets up monitoring and health check endpoints
func SetupMonitoringRoutes(router *gin.Engine, metricsCollector *MetricsCollector, healthChecker *HealthChecker, toolMonitoring *ToolMonitoringSystem) {
	monitoring := router.Group("/monitoring")
	{
		monitoring.GET("/health", healthHandler(healthChecker))
		monitoring.GET("/metrics", metricsHandler(metricsCollector))
		monitoring.GET("/system", systemMetricsHandler(metricsCollector))
		
		// Tool execution monitoring endpoints
		if toolMonitoring != nil && toolMonitoring.IsEnabled() {
			monitoring.GET("/tools/performance", toolPerformanceHandler(toolMonitoring))
			monitoring.GET("/tools/metrics/:toolName", toolSpecificMetricsHandler(toolMonitoring))
			monitoring.GET("/tools/alerts", toolAlertsHandler(toolMonitoring))
		}
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

// toolPerformanceHandler returns tool execution performance metrics
func toolPerformanceHandler(toolMonitoring *ToolMonitoringSystem) gin.HandlerFunc {
	return func(c *gin.Context) {
		metrics := toolMonitoring.GetPerformanceMetrics()
		c.JSON(http.StatusOK, gin.H{
			"tool_performance_metrics": metrics,
			"timestamp": time.Now(),
		})
	}
}

// toolSpecificMetricsHandler returns metrics for a specific tool
func toolSpecificMetricsHandler(toolMonitoring *ToolMonitoringSystem) gin.HandlerFunc {
	return func(c *gin.Context) {
		toolName := c.Param("toolName")
		if toolName == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Tool name is required",
			})
			return
		}

		metrics := toolMonitoring.GetToolMetrics(toolName)
		c.JSON(http.StatusOK, gin.H{
			"tool_metrics": metrics,
			"timestamp": time.Now(),
		})
	}
}

// toolAlertsHandler returns recent tool execution alerts
func toolAlertsHandler(toolMonitoring *ToolMonitoringSystem) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get hours parameter, default to 24 hours
		hoursStr := c.DefaultQuery("hours", "24")
		hours := 24
		if h, err := time.ParseDuration(hoursStr + "h"); err == nil {
			hours = int(h.Hours())
		}

		alerts := toolMonitoring.GetRecentAlerts(hours)
		c.JSON(http.StatusOK, gin.H{
			"alerts": alerts,
			"hours": hours,
			"count": len(alerts),
			"timestamp": time.Now(),
		})
	}
}