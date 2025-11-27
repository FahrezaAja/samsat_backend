package main

import (
	"samsat_backend/config"
	"samsat_backend/routes"

	"github.com/gin-gonic/gin"
)


func main(){
	r := gin.Default()
	config.ConnectDatabase()
	routes.SetupRoutes(r)

	r.Run("0.0.0.0:8080")
}