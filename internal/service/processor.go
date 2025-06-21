package service

import (
	"log"

	"github.com/seansa/rocket-challenge/internal/model"
)

func processMessageWorker(id int, messageChannel <-chan model.IncomingMessage, svc Service) {
	log.Printf("Worker %d started.", id)
	for msg := range messageChannel {
		log.Printf("Worker %d received message for channel %s (msg #%d).", id, msg.Metadata.Channel, msg.Metadata.MessageNumber)
		status, err := svc.ProcessMessage(&msg)
		if err != nil {
			log.Printf("Worker %d ERROR processing message for channel %s (msg #%d): %v", id, msg.Metadata.Channel, msg.Metadata.MessageNumber, err)
		} else {
			log.Printf("Worker %d successfully processed message for channel %s (msg #%d): Status: %s", id, msg.Metadata.Channel, msg.Metadata.MessageNumber, status)
		}
	}
	log.Printf("Worker %d stopped.", id)
}

func StartMessageProcessor(messageChannel <-chan model.IncomingMessage, svc Service, numWorkers int) {
	for i := range numWorkers {
		go processMessageWorker(i+1, messageChannel, svc)
	}
}
