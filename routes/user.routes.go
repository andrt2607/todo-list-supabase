package routes

import (
	"todo-list/controllers"
	"todo-list/middlewares"

	"github.com/gin-gonic/gin"
)

func SetupUserRoutes(router *gin.RouterGroup) {
	router.GET("/self", middlewares.IsAuth, controllers.GetSelf)
	router.GET("/", middlewares.IsAuth, middlewares.IsAdmin, controllers.GetAllUsers)
	router.GET("/:id", middlewares.IsAuth, middlewares.IsAdmin, controllers.GetOneUser)
	router.POST("/", middlewares.IsAuth, middlewares.IsAdmin, controllers.CreateUser)
	router.PATCH("/", middlewares.IsAuth, controllers.EditUser)
	router.PATCH("/password", middlewares.IsAuth, controllers.EditUserPassword)
	router.PATCH("/:id/status", middlewares.IsAuth, middlewares.IsAdmin, controllers.EditUserStatus)
	router.DELETE("/:id", middlewares.IsAuth, middlewares.IsAdmin, controllers.DeleteUser)
}
