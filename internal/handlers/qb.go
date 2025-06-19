package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Status(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Welcome to the Qb API"})
}
