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

// TestProcessMessage_ExistingRocket_DuplicateMessage tests processing a duplicate message.
func TestProcessMessage_ExistingRocket_DuplicateMessage(t *testing.T) {
	mockRepo := new(MockRocketRepository[model.Rocket])
	svc := NewRocketService(mockRepo)

	existingRocket := model.NewRocket("existing-channel-2")
	existingRocket.MessageNumber = 5
	existingRocket.Speed = 500

	testMessage := &model.IncomingMessage{
		Metadata: model.Metadata{
			Channel:       "existing-channel-2",
			MessageNumber: 5, // Duplicate message number
			MessageTime:   time.Now(),
			MessageType:   "RocketSpeedIncreased",
		},
		Message: json.RawMessage(`{"by": 200}`),
	}

	mockRepo.On("Get", "existing-channel-2").Return(existingRocket, nil)
	mockRepo.On("Save", mock.AnythingOfType("model.Rocket")).Return(nil).Run(func(args mock.Arguments) {
		rocket := args.Get(0).(model.Rocket)
		assert.Equal(t, 5, rocket.MessageNumber)
		assert.Equal(t, 700, rocket.Speed)
	})

	status, err := svc.ProcessMessage(testMessage)
	assert.NoError(t, err)
	assert.Equal(t, "re-processed_duplicate", status)
	mockRepo.AssertExpectations(t)
}

// TestProcessMessage_ExistingRocket_OldMessage tests ignoring an old message.
func TestProcessMessage_ExistingRocket_OldMessage(t *testing.T) {
	mockRepo := new(MockRocketRepository[model.Rocket])
	svc := NewRocketService(mockRepo)

	existingRocket := model.NewRocket("existing-channel-3")
	existingRocket.MessageNumber = 10
	existingRocket.Speed = 1000

	testMessage := &model.IncomingMessage{
		Metadata: model.Metadata{
			Channel:       "existing-channel-3",
			MessageNumber: 5, // Older message number
			MessageTime:   time.Now(),
			MessageType:   "RocketSpeedIncreased",
		},
		Message: json.RawMessage(`{"by": 50}`),
	}

	mockRepo.On("Get", "existing-channel-3").Return(existingRocket, nil)
	mockRepo.AssertNotCalled(t, "Save")

	status, err := svc.ProcessMessage(testMessage)
	assert.NoError(t, err)
	assert.Equal(t, "ignoring_old_message", status)
	mockRepo.AssertExpectations(t)
}

// TestProcessMessage_RepoGetError tests repository error during Get.
func TestProcessMessage_RepoGetError(t *testing.T) {
	mockRepo := new(MockRocketRepository[model.Rocket])
	svc := NewRocketService(mockRepo)

	testMessage := &model.IncomingMessage{
		Metadata: model.Metadata{
			Channel:       "error-channel",
			MessageNumber: 1,
			MessageTime:   time.Now(),
			MessageType:   "RocketLaunched",
		},
		Message: json.RawMessage(`{"type": "Falcon-9", "launchSpeed": 100, "mission": "ARTEMIS"}`),
	}

	mockRepo.On("Get", "error-channel").Return(nil, errors.New("simulated repo error"))
	mockRepo.On("Save", mock.AnythingOfType("model.Rocket")).Return(nil)

	status, err := svc.ProcessMessage(testMessage)
	assert.NoError(t, err)
	assert.Equal(t, "processed", status)
	mockRepo.AssertExpectations(t)
}

// TestProcessMessage_RepoSaveError tests repository error during Save.
func TestProcessMessage_RepoSaveError(t *testing.T) {
	mockRepo := new(MockRocketRepository[model.Rocket])
	svc := NewRocketService(mockRepo)

	existingRocket := model.NewRocket("save-error-channel")
	existingRocket.MessageNumber = 0

	testMessage := &model.IncomingMessage{
		Metadata: model.Metadata{
			Channel:       "save-error-channel",
			MessageNumber: 1,
			MessageTime:   time.Now(),
			MessageType:   "RocketLaunched",
		},
		Message: json.RawMessage(`{"type": "Falcon-9", "launchSpeed": 100, "mission": "ARTEMIS"}`),
	}

	mockRepo.On("Get", "save-error-channel").Return(existingRocket, nil)
	mockRepo.On("Save", mock.AnythingOfType("model.Rocket")).Return(errors.New("simulated save error"))

	status, err := svc.ProcessMessage(testMessage)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error saving rocket state")
	assert.Empty(t, status)
	mockRepo.AssertExpectations(t)
}

// TestProcessMessage_ExistingRocket_NewMessage tests processing a new message for an existing rocket.
func TestProcessMessage_ExistingRocket_NewMessage(t *testing.T) {
	mockRepo := new(MockRocketRepository[model.Rocket])
	svc := NewRocketService(mockRepo)

	existingRocket := model.NewRocket("existing-channel-1")
	existingRocket.MessageNumber = 5
	existingRocket.Speed = 500

	testMessage := &model.IncomingMessage{
		Metadata: model.Metadata{
			Channel:       "existing-channel-1",
			MessageNumber: 6,
			MessageTime:   time.Now(),
			MessageType:   "RocketSpeedIncreased",
		},
		Message: json.RawMessage(`{"by": 200}`),
	}

	mockRepo.On("Get", "existing-channel-1").Return(existingRocket, nil)
	mockRepo.On("Save", mock.AnythingOfType("model.Rocket")).Return(nil).Run(func(args mock.Arguments) {
		rocket := args.Get(0).(model.Rocket)
		assert.Equal(t, 6, rocket.MessageNumber)
		assert.Equal(t, 700, rocket.Speed)
	})

	status, err := svc.ProcessMessage(testMessage)
	assert.NoError(t, err)
	assert.Equal(t, "processed", status)
	mockRepo.AssertExpectations(t)
}

// TestGetRocketState_Success tests successful retrieval of a rocket state.
func TestGetRocketState_Success(t *testing.T) {
	mockRepo := new(MockRocketRepository[model.Rocket])
	svc := NewRocketService(mockRepo)

	expectedRocket := model.NewRocket("193270a9-c9cf-404a-8f83-838e71d9ae67")
	expectedRocket.Speed = 500

	mockRepo.On("Get", "193270a9-c9cf-404a-8f83-838e71d9ae67").Return(expectedRocket, nil)

	rocket, err := svc.GetRocketState("193270a9-c9cf-404a-8f83-838e71d9ae67")
	assert.NoError(t, err)
	assert.Equal(t, expectedRocket, rocket)
	mockRepo.AssertExpectations(t)
}

// TestGetRocketState_NotFound tests retrieval of a non-existent rocket.
func TestGetRocketState_NotFound(t *testing.T) {
	mockRepo := new(MockRocketRepository[model.Rocket])
	svc := NewRocketService(mockRepo)

	mockRepo.On("Get", "non-existent").Return(nil, errors.New("not found"))

	rocket, err := svc.GetRocketState("non-existent")
	assert.Error(t, err)
	assert.Zero(t, rocket)
	assert.Contains(t, err.Error(), "not found")
	mockRepo.AssertExpectations(t)
}

// TestGetAllRocketStates_Success tests successful retrieval of all rocket states.
func TestGetAllRocketStates_Success(t *testing.T) {
	mockRepo := new(MockRocketRepository[model.Rocket])
	svc := NewRocketService(mockRepo)

	expectedRockets := []model.Rocket{
		model.NewRocket("193270a9-c9cf-404a-8f83-838e71d9ae67"),
		model.NewRocket("193270a9-c9cf-404a-8f83-838e71d9ae68"),
	}

	mockRepo.On("GetAll").Return(expectedRockets, nil)

	rockets, err := svc.GetAllRocketStates()
	assert.NoError(t, err)
	assert.Equal(t, expectedRockets, rockets)
	mockRepo.AssertExpectations(t)
}

// TestGetAllRocketStates_RepoError tests repository error during GetAllRockets.
func TestGetAllRocketStates_RepoError(t *testing.T) {
	mockRepo := new(MockRocketRepository[model.Rocket])
	svc := NewRocketService(mockRepo)

	mockRepo.On("GetAll").Return(nil, errors.New("repo error"))

	rockets, err := svc.GetAllRocketStates()
	assert.Error(t, err)
	assert.Nil(t, rockets)
	assert.Contains(t, err.Error(), "repo error")
	mockRepo.AssertExpectations(t)
}
