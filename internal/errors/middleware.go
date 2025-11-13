package errors

import (
	"time"

	"github.com/gin-gonic/gin"
)

// ErrorHandler returns a Gin middleware that handles errors added to the context via c.Error().
// It converts APIError types to appropriate JSON responses and wraps unknown errors as internal server errors.
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			requestID, _ := c.Get("request_id")
			reqID, _ := requestID.(string)

			if apiErr, ok := err.Err.(*APIError); ok {
				response := Response{
					Success: false,
					Error: &ErrorInfo{
						Code:      apiErr.Code,
						Message:   apiErr.Message,
						Details:   apiErr.Details,
						Timestamp: time.Now(),
						Path:      c.Request.URL.Path,
						RequestID: reqID,
					},
				}
				c.JSON(apiErr.Status, response)
				return
			}

			response := Response{
				Success: false,
				Error: &ErrorInfo{
					Code:      CodeInternal,
					Message:   "Internal server error",
					Details:   err.Err.Error(),
					Timestamp: time.Now(),
					Path:      c.Request.URL.Path,
					RequestID: reqID,
				},
			}
			c.JSON(500, response)
		}
	}
}
