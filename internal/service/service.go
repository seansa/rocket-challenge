package service

import "github.com/seansa/rocket-challenge/internal/model"

type Service interface {
	ProcessMessage(msg *model.IncomingMessage) (string, error)
	GetRocketState(channel string) (*model.Rocket, error)
	GetAllRocketStates() ([]*model.Rocket, error)
}

type service struct{}

func NewRocketService() Service {
	return &service{}
}

func (s *service) ProcessMessage(msg *model.IncomingMessage) (string, error) {
	panic("Implement me")
}

func (s *service) GetRocketState(channel string) (*model.Rocket, error) {
	panic("Implement me")
}

func (s *service) GetAllRocketStates() ([]*model.Rocket, error) {
	panic("Implement me")
}
