package service

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/seansa/rocket-challenge/internal/model"
	"github.com/seansa/rocket-challenge/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// === MOCKS === //
type MockRocketRepository[T repository.Storable] struct {
	mock.Mock
}

func (m *MockRocketRepository[T]) Get(channel string) (T, error) {
	args := m.Called(channel)
	if args.Get(0) == nil {
		var zero T
		return zero, args.Error(1)
	}
	return args.Get(0).(T), args.Error(1)
}

func (m *MockRocketRepository[T]) GetAll() ([]T, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]T), args.Error(1)
}

func (m *MockRocketRepository[T]) Save(rocket T) error {
	args := m.Called(rocket)
	return args.Error(0)
}

// === END MOCKS === //

func TestNewRocketService(t *testing.T) {
	mockRepo := new(MockRocketRepository[model.Rocket])
	svc := NewRocketService(mockRepo)
	assert.NotNil(t, svc)
}

// TestProcessMessage_NewRocket tests processing a message for a new rocket.
func TestProcessMessage_NewRocket(t *testing.T) {
	mockRepo := new(MockRocketRepository[model.Rocket])
	svc := NewRocketService(mockRepo)

	testMessage := &model.IncomingMessage{
		Metadata: model.Metadata{
			Channel:       "193270a9-c9cf-404a-8f83-838e71d9ae67",
			MessageNumber: 1,
			MessageTime:   time.Now(),
			MessageType:   "RocketLaunched",
		},
		Message: json.RawMessage(`{"type": "Falcon-9", "launchSpeed": 100, "mission": "ARTEMIS"}`),
	}

	mockRepo.On("Get", "193270a9-c9cf-404a-8f83-838e71d9ae67").Return(nil, errors.New("not found"))
	mockRepo.On("Save", mock.AnythingOfType("model.Rocket")).Return(nil).Run(func(args mock.Arguments) {
		rocket := args.Get(0).(model.Rocket)
		assert.Equal(t, "193270a9-c9cf-404a-8f83-838e71d9ae67", rocket.Channel)
		assert.Equal(t, 1, rocket.MessageNumber) // Message number should be updated
		assert.Equal(t, "Falcon-9", rocket.Type)
		assert.Equal(t, 100, rocket.Speed)
	})

	status, err := svc.ProcessMessage(testMessage)
	assert.NoError(t, err)
	assert.Equal(t, "processed", status)
	mockRepo.AssertExpectations(t)
}
