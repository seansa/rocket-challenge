package cmd

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
)

func Run() {
	router := gin.Default()

	port := getOrDefault("PORT", ":8088")

	log.Printf("Server listening on http://localhost%s", port)
	log.Fatal(router.Run(port))
}

func getOrDefault(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value

}
