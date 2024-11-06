package main

import (
	"os"

	"todo-list/db"
	"todo-list/middlewares"
	"todo-list/routes"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	_ "github.com/joho/godotenv/autoload"
)

func main() {

	db.InitPostgres() // init db (postgres)

	is_production := os.Getenv("PRODUCTION")
	if is_production == "true" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()
	r.Use(middlewares.CORS())

	// api routing
	apiV1 := r.Group("/api/v1")

	validate := validator.New()
	routes.SetupAuthRoutes(apiV1.Group("/auth"))
	routes.SetupUserRoutes(apiV1.Group("/users"))
	routes.SetupActivityRoutes(apiV1.Group("/activities"), validate)

	// start
	port := os.Getenv("PORT")
	if port == "" {
		port = "8888"
	}

	err := r.Run(":" + port)
	if err != nil {
		panic(err)
	}

}
