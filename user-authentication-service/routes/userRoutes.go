package routes

import (
	controller "github.com/vikjdk7/Algotrading-GoLang-Rest/user-authentication-service/controllers"

	"github.com/gin-gonic/gin"
)

//UserRoutes function
func UserRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.POST("/UserService/api/v1/signupemail", controller.SignUpEmailSend())
	incomingRoutes.POST("/UserService/api/v1/signup", controller.SignUp())

	incomingRoutes.POST("/UserService/api/v1/login", controller.Login())

	incomingRoutes.POST("/UserService/api/v1/resetpassword", controller.ResetPassword())

	incomingRoutes.POST("/UserService/api/v1/forgotpasswordemail", controller.ForgotPasswordEmailSend())
	incomingRoutes.POST("/UserService/api/v1/forgotpasswordreset", controller.ForgotPasswordReset())

	incomingRoutes.PUT("/UserService/api/v1/userprofile", controller.UserProfile())
	incomingRoutes.GET("/UserService/api/v1/userprofile", controller.GetUserProfile())
	incomingRoutes.DELETE("/UserService/api/v1/deleteuser", controller.DeleteUser())
}
