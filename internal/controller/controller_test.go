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

func (m *MockRocketService) GetRocketState(channel string) (*model.Rocket, error) {
	args := m.Called(channel)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Rocket), args.Error(1)
}

func (m *MockRocketService) GetAllRocketStates() ([]*model.Rocket, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Rocket), args.Error(1)
}

// === END MOCKS === //

// setupRouter configures a Gin router for testing.
func setupRouter(mockService *MockRocketService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	r.RedirectTrailingSlash = false

	controller := NewRocketController(mockService)
	r.POST("/messages", controller.ReceiveMessageHandler)
	r.GET("/rockets", controller.GetAllRocketsHandler)
	r.GET("/rockets/:channel", controller.GetRocketStateHandler)
	return r
}

// TestReceiveMessageHandler_Success tests successful message handling.
func TestReceiveMessageHandler_Success(t *testing.T) {
	mockService := new(MockRocketService)
	router := setupRouter(mockService)

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

	mockService.On("ProcessMessage", mock.Anything).Return("processed", nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/messages", bytes.NewBuffer(msgBytes))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"status":"processed"`)
	assert.Contains(t, w.Body.String(), `"channel":"193270a9-c9cf-404a-8f83-838e71d9ae67"`)
	mockService.AssertExpectations(t)
}

// TestReceiveMessageHandler_InvalidJSON tests an invalid JSON payload.
func TestReceiveMessageHandler_InvalidJSON(t *testing.T) {
	mockService := new(MockRocketService)
	router := setupRouter(mockService)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/messages", bytes.NewBuffer([]byte(`{"invalid json`)))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), `"error":"Invalid JSON or empty request body"`)
	mockService.AssertNotCalled(t, "ProcessMessage")
}

// TestReceiveMessageHandler_MissingMetadata tests missing essential metadata.
func TestReceiveMessageHandler_MissingMetadata(t *testing.T) {
	mockService := new(MockRocketService)
	router := setupRouter(mockService)

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
	mockService.AssertNotCalled(t, "ProcessMessage")
}

// TestReceiveMessageHandler_ServiceError tests when the service returns an error.
func TestReceiveMessageHandler_ServiceError(t *testing.T) {
	mockService := new(MockRocketService)
	router := setupRouter(mockService)

	testMessage := model.IncomingMessage{
		Metadata: model.Metadata{
			Channel:       "channel-err",
			MessageNumber: 1,
			MessageTime:   time.Now(),
			MessageType:   "RocketLaunched",
		},
		Message: json.RawMessage(`{"type": "Falcon-9"}`),
	}
	msgBytes, _ := json.Marshal(testMessage)

	mockService.On("ProcessMessage", mock.Anything).Return("", errors.New("simulated service error"))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/messages", bytes.NewBuffer(msgBytes))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), `"error":"Error processing message for channel channel-err"`)
	mockService.AssertExpectations(t)
}
