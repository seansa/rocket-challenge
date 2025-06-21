package model

import (
	"encoding/json"
	"time"
)

// IncomingMessage represents the full structure of an incoming JSON message.
type IncomingMessage struct {
	Metadata Metadata        `json:"metadata"`
	Message  json.RawMessage `json:"message"`
}

// Metadata represents the 'metadata' section of a rocket message.
type Metadata struct {
	Channel       string      `json:"channel" binding:"required"`
	MessageNumber int         `json:"messageNumber" binding:"required"`
	MessageTime   time.Time   `json:"messageTime" binding:"required"`
	MessageType   MessageType `json:"messageType" binding:"required"`
}

// LaunchedMessage represents the 'message' section for the RocketLaunched type.
type LaunchedMessage struct {
	Type        string `json:"type"`
	LaunchSpeed int    `json:"launchSpeed"`
	Mission     string `json:"mission"`
}

// SpeedChangedMessage represents the 'message' section for speed increase/decrease.
type SpeedChangedMessage struct {
	By int `json:"by"`
}

// RocketExplodedMessage represents the 'message' section for the RocketExploded event.
type RocketExplodedMessage struct {
	Reason string `json:"reason"`
}

// MissionChangedMessage represents the 'message' section for the RocketMissionChanged event.
type MissionChangedMessage struct {
	NewMission string `json:"newMission"`
}
