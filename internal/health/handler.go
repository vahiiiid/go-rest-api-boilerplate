package health

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

// Health godoc
// @Summary      Basic health check
// @Description  Check if the application is running
// @Tags         Health
// @Accept       json
// @Produce      json
// @Success      200  {object}  HealthResponse
// @Router       /health [get]
func (h *Handler) Health(c *gin.Context) {
	ctx := c.Request.Context()
	response := h.service.GetHealth(ctx)
	c.JSON(http.StatusOK, response)
}

// Live godoc
// @Summary      Liveness probe
// @Description  Check if the application is alive (not deadlocked)
// @Tags         Health
// @Accept       json
// @Produce      json
// @Success      200  {object}  HealthResponse
// @Router       /health/live [get]
func (h *Handler) Live(c *gin.Context) {
	ctx := c.Request.Context()
	response := h.service.GetLiveness(ctx)
	c.JSON(http.StatusOK, response)
}

// Ready godoc
// @Summary      Readiness probe
// @Description  Check if the application and its dependencies are ready to serve traffic
// @Tags         Health
// @Accept       json
// @Produce      json
// @Success      200  {object}  HealthResponse  "Service is ready"
// @Success      503  {object}  HealthResponse  "Service is not ready"
// @Router       /health/ready [get]
func (h *Handler) Ready(c *gin.Context) {
	ctx := c.Request.Context()
	response := h.service.GetReadiness(ctx)

	statusCode := http.StatusOK
	if response.Status == StatusUnhealthy {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, response)
}
