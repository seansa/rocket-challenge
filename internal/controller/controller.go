package controller

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/seansa/rocket-challenge/internal/model"
	"github.com/seansa/rocket-challenge/internal/service"
)

type rocketController struct {
	service service.Service
}

func NewRocketController(service service.Service) *rocketController {
	return &rocketController{
		service: service,
	}
}

func (c *rocketController) MessageHandler(ctx *gin.Context) {
	var msg model.IncomingMessage

	if err := ctx.ShouldBindJSON(&msg); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON or empty request body", "details": err.Error()})
		return
	}

	resp, err := c.service.ProcessMessage(&msg)
	if err != nil {
		log.Printf("Error processing message in service: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error processing message for channel %s", msg.Metadata.Channel), "details": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": resp, "channel": msg.Metadata.Channel})

}

func (c *rocketController) GetAllRocketsHandler(ctx *gin.Context) {
	rockets, err := c.service.GetAllRocketStates()
	if err != nil {
		log.Printf("Error getting all rockets from service: %+v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error while fetching rockets", "details": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, rockets)
}

func (c *rocketController) GetRocketStateHandler(ctx *gin.Context) {
	ctx.JSON(200, gin.H{
		"message": "Rocket state retrieved successfully",
	})
}
