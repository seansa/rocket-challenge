package cmd

import (
	"github.com/gin-gonic/gin"
	"github.com/seansa/rocket-challenge/internal/controller"
	"github.com/seansa/rocket-challenge/internal/service"
)

func mapUrls(r *gin.Engine) {
	service := service.NewRocketService()
	controller := controller.NewRocketController(service)

	r.POST("/messages", controller.MessageHandler)
	r.GET("/rockets", controller.GetAllRocketsHandler)
	r.GET("/rockets/:channel", controller.GetRocketStateHandler)

}
