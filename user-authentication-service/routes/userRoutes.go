package routes

import (
	controller "github.com/vikjdk7/Algotrading-GoLang-Rest/user-authentication-service/controllers"

	"github.com/gin-gonic/gin"
)

//UserRoutes function
func UserRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.POST("/UserService/api/v1/signup", controller.SignUp())
	incomingRoutes.POST("/UserService/api/v1/login", controller.Login())
}
