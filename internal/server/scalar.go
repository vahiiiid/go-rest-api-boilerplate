package server

import (
	"net/http"

	"github.com/gin-gonic/gin"

	scalar "github.com/MarceloPetrucio/go-scalar-api-reference"
)

func scalarHandler(ctx *gin.Context) {
	htmlContent, err := scalar.ApiReferenceHTML(&scalar.Options{
		SpecURL: "/swagger/doc.json",
		CustomOptions: scalar.CustomOptions{
			PageTitle: "Afghanistan Base Project Structure",
		},
		DarkMode: true,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate API reference",
		})
		return
	}

	ctx.Data(http.StatusOK, "text/html; charset=utf-8", []byte(htmlContent))
}
