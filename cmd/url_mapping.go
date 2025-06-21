package cmd

import (
	"github.com/gin-gonic/gin"
	"github.com/seansa/rocket-challenge/internal/controller"
	"github.com/seansa/rocket-challenge/internal/model"
	"github.com/seansa/rocket-challenge/internal/repository"
	"github.com/seansa/rocket-challenge/internal/service"
)

func mapUrls(r *gin.Engine) {
	repository := repository.NewRepository[model.Rocket]()
	service := service.NewRocketService(repository)
	controller := controller.NewRocketController(service)

	r.POST("/messages", controller.MessageHandler)
	r.GET("/rockets", controller.GetAllRocketsHandler)
	r.GET("/rockets/:channel", controller.GetRocketStateHandler)

}
