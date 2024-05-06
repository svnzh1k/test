package main

import (
	"rest/api/controllers"
	"rest/api/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	routes.Setup(router)
	controllers.Init()
	router.Run(":8080")
}
