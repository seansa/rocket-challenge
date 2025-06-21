package controller

import "github.com/gin-gonic/gin"

type rocketController struct {
}

func NewRocketController() *rocketController {
	return &rocketController{}
}

func (c *rocketController) ReceiveMessageHandler(ctx *gin.Context) {
	ctx.JSON(200, gin.H{
		"message": "Message received successfully",
	})
}

func (c *rocketController) GetAllRocketsHandler(ctx *gin.Context) {
	ctx.JSON(200, gin.H{
		"message": "All rockets retrieved successfully",
	})
}

func (c *rocketController) GetRocketStateHandler(ctx *gin.Context) {
	ctx.JSON(200, gin.H{
		"message": "Rocket state retrieved successfully",
	})
}
