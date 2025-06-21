package controller

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/seansa/rocket-challenge/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// === MOCKS === //
type MockRocketService struct {
	mock.Mock
}

func (m *MockRocketService) ProcessMessage(msg *model.IncomingMessage) (string, error) {
	args := m.Called(msg)
	return args.String(0), args.Error(1)
}

func (m *MockRocketService) GetRocketState(channel string) (model.Rocket, error) {
	args := m.Called(channel)
	if args.Get(0) == nil {
		return model.Rocket{}, args.Error(1)
	}
	return args.Get(0).(model.Rocket), args.Error(1)
}

func (m *MockRocketService) GetAllRocketStates() ([]model.Rocket, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Rocket), args.Error(1)
}

// === END MOCKS === //

func setupRouter(mockService *MockRocketService, messageChannel chan model.IncomingMessage) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	r.RedirectTrailingSlash = false

	controller := NewRocketController(mockService, messageChannel)
	r.POST("/messages", controller.MessageHandler)
	r.GET("/rockets", controller.GetAllRocketsHandler)
	r.GET("/rockets/:channel", controller.GetRocketStateHandler)
	return r
}

// TestReceiveMessageHandler_Success tests successful message handling by sending to channel.
func TestReceiveMessageHandler_Success(t *testing.T) {
	mockService := new(MockRocketService)
	testMessageChannel := make(chan model.IncomingMessage, 1)
	router := setupRouter(mockService, testMessageChannel)

	testMessage := model.IncomingMessage{
		Metadata: model.Metadata{
			Channel:       "193270a9-c9cf-404a-8f83-838e71d9ae67",
			MessageNumber: 1,
			MessageTime:   time.Now(),
			MessageType:   "RocketLaunched",
		},
		Message: json.RawMessage(`{"type": "Falcon-9", "launchSpeed": 500, "mission": "ARTEMIS"}`),
	}
	msgBytes, _ := json.Marshal(testMessage)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/messages", bytes.NewBuffer(msgBytes))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusAccepted, w.Code) // Expect 202 Accepted
	assert.Contains(t, w.Body.String(), `"status":"accepted_for_processing"`)
	assert.Contains(t, w.Body.String(), `"channel":"193270a9-c9cf-404a-8f83-838e71d9ae67"`)

	select {
	case receivedMsg := <-testMessageChannel:
		assert.Equal(t, testMessage.Metadata.Channel, receivedMsg.Metadata.Channel)
		assert.Equal(t, testMessage.Metadata.MessageNumber, receivedMsg.Metadata.MessageNumber)
		expectedMsgBytes, _ := json.Marshal(testMessage.Message)
		actualMsgBytes, _ := json.Marshal(receivedMsg.Message)
		assert.JSONEq(t, string(expectedMsgBytes), string(actualMsgBytes))
	case <-time.After(time.Millisecond * 100):
		t.Fatal("Message not received on channel within timeout")
	}

	mockService.AssertNotCalled(t, "ProcessMessage")
	close(testMessageChannel)
}

// TestReceiveMessageHandler_InvalidJSON tests an invalid JSON payload.
func TestReceiveMessageHandler_InvalidJSON(t *testing.T) {
	mockService := new(MockRocketService)
	testMessageChannel := make(chan model.IncomingMessage, 1)
	router := setupRouter(mockService, testMessageChannel)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/messages", bytes.NewBuffer([]byte(`{"invalid json`)))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), `"error":"Invalid JSON or empty request body"`)

	select {
	case <-testMessageChannel:
		t.Fatal("Unexpected message received on channel")
	default:
	}
	mockService.AssertNotCalled(t, "ProcessMessage")
	close(testMessageChannel)
}

// TestReceiveMessageHandler_MissingMetadata tests missing essential metadata.
func TestReceiveMessageHandler_MissingMetadata(t *testing.T) {
	mockService := new(MockRocketService)
	testMessageChannel := make(chan model.IncomingMessage, 1)
	router := setupRouter(mockService, testMessageChannel)

	testMessage := model.IncomingMessage{
		Metadata: model.Metadata{
			Channel:       "", // Missing Channel
			MessageNumber: 1,
			MessageTime:   time.Now(),
			MessageType:   "RocketLaunched",
		},
		Message: json.RawMessage(`{"type": "Falcon-9"}`),
	}
	msgBytes, _ := json.Marshal(testMessage)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/messages", bytes.NewBuffer(msgBytes))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), `"error":"Invalid JSON or empty request body"`)

	select {
	case <-testMessageChannel:
		t.Fatal("Unexpected message received on channel")
	default:
	}
	mockService.AssertNotCalled(t, "ProcessMessage")
	close(testMessageChannel)
}

// TestReceiveMessageHandler_QueueFull tests when the message channel is full.
func TestReceiveMessageHandler_QueueFull(t *testing.T) {
	mockService := new(MockRocketService)
	testMessageChannel := make(chan model.IncomingMessage)
	router := setupRouter(mockService, testMessageChannel)

	testMessage := model.IncomingMessage{
		Metadata: model.Metadata{
			Channel:       "channel-full",
			MessageNumber: 1,
			MessageTime:   time.Now(),
			MessageType:   "RocketLaunched",
		},
		Message: json.RawMessage(`{"type": "Falcon-9", "launchSpeed": 500, "mission": "ARTEMIS"}`),
	}
	msgBytes, _ := json.Marshal(testMessage)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/messages", bytes.NewBuffer(msgBytes))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code) // Expect 503 Service Unavailable
	assert.Contains(t, w.Body.String(), `"error":"Message queue full, please try again later"`)

	select { // Ensure no message was sent to the channel
	case <-testMessageChannel:
		t.Fatal("Unexpected message received on channel")
	default:
		// No message, as expected
	}
	mockService.AssertNotCalled(t, "ProcessMessage")
	close(testMessageChannel)
}

// TestGetAllRocketsHandler_Success tests successful retrieval of all rockets.
func TestGetAllRocketsHandler_Success(t *testing.T) {
	mockService := new(MockRocketService)
	testMessageChannel := make(chan model.IncomingMessage)
	router := setupRouter(mockService, testMessageChannel)

	expectedRockets := []model.Rocket{
		{Channel: "193270a9-c9cf-404a-8f83-838e71d9ae67", Type: "Falcon-9", Speed: 500, Mission: "ARTEMIS"},
		{Channel: "193270a9-c9cf-404a-8f83-838e71d9ae68", Type: "Falcon-8", Speed: 1500, Mission: "ARTEMIS2"},
	}

	mockService.On("GetAllRocketStates").Return(expectedRockets, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/rockets", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var actualRockets []model.Rocket
	err := json.Unmarshal(w.Body.Bytes(), &actualRockets)
	assert.NoError(t, err)
	assert.Len(t, actualRockets, 2)
	assert.Equal(t, expectedRockets[0].Channel, actualRockets[0].Channel)
	mockService.AssertExpectations(t)
	close(testMessageChannel)
}

// TestGetAllRocketsHandler_ServiceError tests when the GetAll service returns an error.
func TestGetAllRocketsHandler_ServiceError(t *testing.T) {
	mockService := new(MockRocketService)
	testMessageChannel := make(chan model.IncomingMessage)
	router := setupRouter(mockService, testMessageChannel)

	mockService.On("GetAllRocketStates").Return(nil, errors.New("foo bar error"))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/rockets", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), `"error":"Error while fetching rockets"`)
	mockService.AssertExpectations(t)
	close(testMessageChannel)
}

// TestGetRocketStateHandler_Success tests successful retrieval of a single rocket.
func TestGetRocketStateHandler_Success(t *testing.T) {
	mockService := new(MockRocketService)
	testMessageChannel := make(chan model.IncomingMessage)
	router := setupRouter(mockService, testMessageChannel)

	expectedRocket := model.Rocket{Channel: "193270a9-c9cf-404a-8f83-838e71d9ae67", Type: "Falcon-9", Speed: 500, Mission: "ARTEMIS"}

	mockService.On("GetRocketState", "193270a9-c9cf-404a-8f83-838e71d9ae67").Return(expectedRocket, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/rockets/193270a9-c9cf-404a-8f83-838e71d9ae67", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var actualRocket model.Rocket
	err := json.Unmarshal(w.Body.Bytes(), &actualRocket)
	assert.NoError(t, err)
	assert.Equal(t, expectedRocket.Channel, actualRocket.Channel)
	mockService.AssertExpectations(t)
	close(testMessageChannel)
}

// TestGetRocketStateHandler_NotFound tests when the rocket is not found.
func TestGetRocketStateHandler_NotFound(t *testing.T) {
	mockService := new(MockRocketService)
	testMessageChannel := make(chan model.IncomingMessage)
	router := setupRouter(mockService, testMessageChannel)

	mockService.On("GetRocketState", "non-existent-channel").Return(nil, errors.New("rocket with channel non-existent-channel not found"))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/rockets/non-existent-channel", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), `"error":"Rocket not found"`)
	mockService.AssertExpectations(t)
	close(testMessageChannel)
}

// TestGetRocketStateHandler_ServiceError tests when the Get service returns a generic error.
func TestGetRocketStateHandler_ServiceError(t *testing.T) {
	mockService := new(MockRocketService)
	testMessageChannel := make(chan model.IncomingMessage)
	router := setupRouter(mockService, testMessageChannel)

	mockService.On("GetRocketState", "errChannel").Return(nil, errors.New("internal repository error"))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/rockets/errChannel", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), `"error":"Error while fetching rocket errChannel"`)
	mockService.AssertExpectations(t)
	close(testMessageChannel)
}
