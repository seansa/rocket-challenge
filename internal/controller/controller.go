package controller

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/seansa/rocket-challenge/internal/model"
	"github.com/seansa/rocket-challenge/internal/service"
)

type RocketController struct {
	service        service.Service
	messageChannel chan<- model.IncomingMessage
}

func NewRocketController(service service.Service, msgChan chan<- model.IncomingMessage) *RocketController {
	return &RocketController{
		service:        service,
		messageChannel: msgChan,
	}
}

// MessageHandler handles incoming POST requests to the /messages endpoint.
// @Summary Receive rocket message
// @Description Processes an incoming rocket state message. Handles out-of-order and duplicate messages.
// @Tags messages
// @Accept json
// @Produce json
// @Param message body model.IncomingMessage true "Rocket message payload"
// @Success 200 {object} map[string]string "Status of message processing"
// @Failure 400 {object} map[string]string "Invalid JSON or bad request"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /messages [post]
func (c *RocketController) MessageHandler(ctx *gin.Context) {
	var msg model.IncomingMessage

	if err := ctx.ShouldBindJSON(&msg); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON or empty request body", "details": err.Error()})
		return
	}

	select {
	case c.messageChannel <- msg:
		log.Printf("Message for channel %s (msg #%d) accepted for processing.", msg.Metadata.Channel, msg.Metadata.MessageNumber)
		// Return 202 Accepted, indicating the request has been accepted for processing.
		ctx.JSON(http.StatusAccepted, gin.H{"status": "accepted_for_processing", "channel": msg.Metadata.Channel})
	default:
		// If the channel is full, respond with Service Unavailable (503).
		log.Printf("Message for channel %s (msg #%d) rejected: message queue full.", msg.Metadata.Channel, msg.Metadata.MessageNumber)
		ctx.JSON(http.StatusServiceUnavailable, gin.H{"error": "Message queue full, please try again later"})
	}

}

// GetAllRocketsHandler handles GET requests to the /rockets endpoint.
// @Summary Get all rocket states
// @Description Returns a list of the current states of all rockets in the system, sorted by channel ID.
// @Tags rockets
// @Produce json
// @Success 200 {array} model.Rocket "List of all rockets"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /rockets [get]
func (c *RocketController) GetAllRocketsHandler(ctx *gin.Context) {
	rockets, err := c.service.GetAllRocketStates()
	if err != nil {
		log.Printf("Error getting all rockets from service: %+v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error while fetching rockets", "details": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, rockets)
}

// GetRocketStateHandler handles GET requests to the /rockets/{channel} endpoint.
// @Summary Get a single rocket state
// @Description Returns the current state of a specific rocket by its channel ID.
// @Tags rockets
// @Produce json
// @Param channel path string true "Rocket Channel ID"
// @Success 200 {object} model.Rocket "Current state of the rocket"
// @Failure 400 {object} map[string]string "Missing rocket channel ID"
// @Failure 404 {object} map[string]string "Rocket not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /rockets/{channel} [get]
func (c *RocketController) GetRocketStateHandler(ctx *gin.Context) {
	channel := ctx.Param("channel")

	rocket, err := c.service.GetRocketState(channel)
	if err != nil {
		//TODO: create a custom error type for better error handling
		if strings.Contains(err.Error(), "not found") {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Rocket not found", "channel": channel})
		} else {
			log.Printf("Error getting rocket state %s from service: %v", channel, err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error while fetching rocket %s", channel), "details": err.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, rocket)
}
