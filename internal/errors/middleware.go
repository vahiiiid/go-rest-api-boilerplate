package errors

import (
	"github.com/gin-gonic/gin"
)

// ErrorHandler returns a Gin middleware that handles errors added to the context via c.Error().
// It converts APIError types to appropriate JSON responses and wraps unknown errors as internal server errors.
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last()

			if apiErr, ok := err.Err.(*APIError); ok {
				c.JSON(apiErr.Status, apiErr)
				return
			}

			c.JSON(500, InternalServerError(err.Err))
		}
	}
}
