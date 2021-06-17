package main

import (
	"os"

	routes "github.com/vikjdk7/Algotrading-GoLang-Rest/user-authentication-service/routes"

	"github.com/gin-gonic/gin"
	_ "github.com/heroku/x/hmetrics/onload"
)

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		port = "3000"
	}

	router := gin.New()
	router.Use(gin.Logger())
	routes.UserRoutes(router)

	router.Run(":" + port)
}
