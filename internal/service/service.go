package service

import (
	"fmt"
	"log"

	"github.com/seansa/rocket-challenge/internal/model"
	"github.com/seansa/rocket-challenge/internal/repository"
)

type Service interface {
	ProcessMessage(msg *model.IncomingMessage) (string, error)
	GetRocketState(channel string) (*model.Rocket, error)
	GetAllRocketStates() ([]*model.Rocket, error)
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

	rocket, err := s.repo.Get(channel)
	if err != nil {
		// If the rocket is not found, assume it's a new rocket.
		rocket = model.NewRocket(channel)
		log.Printf("New rocket registered in service: %s", channel)
	}

	if err := rocket.UpdateState(msg.Metadata.MessageType, msg.Message); err != nil {
		return "", fmt.Errorf("error updating rocket state %s: %w", channel, err)
	}
	rocket.MessageNumber = msg.Metadata.MessageNumber
	rocket.MessageTime = msg.Metadata.MessageTime

	if err := s.repo.Save(rocket); err != nil {
		return "", fmt.Errorf("error saving rocket state %s: %w", channel, err)
	}

	return "processed", nil
}

func (s *service) GetRocketState(channel string) (*model.Rocket, error) {
	rocket, err := s.repo.Get(channel)
	if err != nil {
		return nil, err
	}
	log.Printf("Returning state for rocket %s.", channel)
	return &rocket, nil
}

func (s *service) GetAllRocketStates() ([]*model.Rocket, error) {
	panic("Implement me")
}
