package cmd

import (
	"github.com/gin-gonic/gin"
	"github.com/seansa/rocket-challenge/internal/controller"
)

func mapUrls(r *gin.Engine) {
	rocketController := controller.NewRocketController()

	r.POST("/messages", rocketController.ReceiveMessageHandler)
	r.GET("/rockets", rocketController.GetAllRocketsHandler)
	r.GET("/rockets/:channel", rocketController.GetRocketStateHandler)

}
