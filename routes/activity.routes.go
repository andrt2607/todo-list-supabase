package routes

import (
	"todo-list/controllers"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func SetupActivityRoutes(router *gin.RouterGroup, validate *validator.Validate) {

	activityController := controllers.NewActivityController(validate)

	router.GET("/", activityController.GetActivities)
	router.POST("/", activityController.CreateActivity)
	router.PUT("/:id", activityController.UpdateActivity)
	router.DELETE("/:id", activityController.DeleteActivity)
	router.GET("/export-excel", activityController.ExportExcelActivities)
	router.GET("/export-pdf", activityController.ExportPDFActivities)
}
