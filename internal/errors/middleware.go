package errors

import (
	"github.com/gin-gonic/gin"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last()

			// Handle APIError
			if apiErr, ok := err.Err.(*APIError); ok {
				c.JSON(apiErr.Status, apiErr)
				return
			}

			// Handle unknown errors
			c.JSON(500, InternalServerError(err.Err))
		}
	}
}
