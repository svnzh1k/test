package routes

import (
	"rest/api/controllers"

	"github.com/gin-gonic/gin"
)

func Setup(router *gin.Engine) {
	router.POST("/auth/signup", controllers.HandleSignup)
	router.POST("/auth/login", controllers.HandleLogin)
	router.POST("/admin/menu", controllers.AddToMenu)
	router.DELETE("/admin/menu/:id", controllers.RemoveFromMenu)
	router.PATCH("/admin/:id", controllers.UpdateStatus)
	router.POST("/menu/:id", controllers.PlaceOrder)
	router.GET("/menu", controllers.ShowMenu)
	router.GET("/admin/revenue", controllers.GetRevenue)
	router.GET("/documentation", controllers.GetDocumentation)
}
