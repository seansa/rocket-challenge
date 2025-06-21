package service

import (
	"fmt"
	"log"

	"github.com/seansa/rocket-challenge/internal/model"
	"github.com/seansa/rocket-challenge/internal/repository"
)

type Service interface {
	ProcessMessage(msg *model.IncomingMessage) (string, error)
	GetRocketState(channel string) (model.Rocket, error)
	GetAllRocketStates() ([]model.Rocket, error)
}

type service struct {
	repo repository.Repository[model.Rocket]
}

func NewRocketService(repo repository.Repository[model.Rocket]) Service {
	return &service{
		repo: repo,
	}
}

func (s *service) ProcessMessage(msg *model.IncomingMessage) (string, error) {
	channel := msg.Metadata.Channel
	incomingMessageNumber := msg.Metadata.MessageNumber
	incomingMessageType := msg.Metadata.MessageType
	incomingMessageData := msg.Message

	savedRocket, err := s.repo.Get(channel)
	if err != nil {
		// If the rocket is not found, assume it's a new rocket.
		savedRocket = model.NewRocket(channel)
		log.Printf("New rocket registered in service: %s", channel)
	}

	statusMsg := "ignoring_old_message"
	stateChanged := false

	// Primary logic for handling out-of-order and duplicate messages:
	// We only process a message if its messageNumber is strictly greater than
	// the last messageNumber we've seen for this rocket.
	// If it's the same messageNumber, we re-process it (idempotency for duplicates).
	// If it's an older message (lower messageNumber), we ignore it.
	if incomingMessageNumber > savedRocket.MessageNumber {
		if err := savedRocket.UpdateState(incomingMessageType, incomingMessageData); err != nil {
			return "", fmt.Errorf("error updating rocket state %s: %w", channel, err)
		}
		savedRocket.MessageNumber = incomingMessageNumber
		savedRocket.MessageTime = msg.Metadata.MessageTime
		statusMsg = "processed"
		stateChanged = true
	} else if incomingMessageNumber == savedRocket.MessageNumber {
		// If a duplicate of the current latest message arrives, re-process it.
		// This ensures idempotency for at-least-once delivery.
		// or we can ignore it if we want to avoid re-processing because we already processed it.
		// Here we assume that re-processing is safe and ensure at-least-once delivery
		log.Printf("Re-processing duplicate message %d for channel %s.", incomingMessageNumber, channel)
		if err := savedRocket.UpdateState(incomingMessageType, incomingMessageData); err != nil {
			return "", fmt.Errorf("error re-processing rocket state %s: %w", channel, err)
		}
		statusMsg = "re-processed_duplicate"
		stateChanged = true // We can change to false if we want to avoid re-processing
	} else {
		log.Printf("Ignoring old message %d for channel %s (current last: %d).", incomingMessageNumber, channel, savedRocket.MessageNumber)
	}

	if stateChanged {
		if err := s.repo.Save(savedRocket); err != nil {
			return "", fmt.Errorf("error saving rocket state %s: %w", channel, err)
		}
	}

	return statusMsg, nil
}

func (s *service) GetRocketState(channel string) (model.Rocket, error) {
	rocket, err := s.repo.Get(channel)
	if err != nil {
		return model.Rocket{}, err
	}
	log.Printf("Returning state for rocket %s.", channel)
	return rocket, nil
}

func (s *service) GetAllRocketStates() ([]model.Rocket, error) {
	rockets, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}
	log.Printf("Returning list of %d rockets.", len(rockets))
	return rockets, nil
}
