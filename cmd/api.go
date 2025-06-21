package cmd

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/seansa/rocket-challenge/internal/controller"
	"github.com/seansa/rocket-challenge/internal/model"
	"github.com/seansa/rocket-challenge/internal/repository"
	"github.com/seansa/rocket-challenge/internal/service"
)

var (
	repo           repository.Repository[model.Rocket]
	srv            service.Service
	ctrl           *controller.RocketController
	messageChannel = make(chan model.IncomingMessage, 1000)
	numWorkers     = 5
)

func Run() {
	port := getOrDefault("PORT", ":8088")

	setupDependencies()
	setupWorkers()
	r := setupRoutes()

	log.Printf("Server listening on http://localhost%s", port)
	log.Fatal(r.Run(port))
}

func setupDependencies() {
	repo = repository.NewRepository[model.Rocket]()
	srv = service.NewRocketService(repo)
	ctrl = controller.NewRocketController(srv, messageChannel)
}

func setupWorkers() {
	service.StartMessageProcessor(messageChannel, srv, numWorkers)
	log.Printf("Started %d message processing workers.", numWorkers)
}

func setupRoutes() *gin.Engine {
	r := gin.Default()
	r.RedirectTrailingSlash = false

	r.POST("/messages", ctrl.MessageHandler)
	r.GET("/rockets", ctrl.GetAllRocketsHandler)
	r.GET("/rockets/:channel", ctrl.GetRocketStateHandler)

	return r
}

func getOrDefault(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
