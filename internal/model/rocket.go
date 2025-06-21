package model

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
)

type MessageType string

const (
	RocketLaunched       MessageType = "RocketLaunched"
	RocketSpeedIncreased MessageType = "RocketSpeedIncreased"
	RocketSpeedDecreased MessageType = "RocketSpeedDecreased"
	RocketExploded       MessageType = "RocketExploded"
	RocketMissionChanged MessageType = "RocketMissionChanged"
)

const (
	Aborted string = "ABORTED"
)

// Rocket represents the current state of a rocket.
type Rocket struct {
	Channel         string    `json:"channel"`
	Type            string    `json:"type,omitempty"`
	Speed           int       `json:"speed"`
	Mission         string    `json:"mission,omitempty"`
	Exploded        bool      `json:"exploded"`
	ExplosionReason string    `json:"explosionReason,omitempty"`
	MessageNumber   int       `json:"-"`
	MessageTime     time.Time `json:"-"`
}

// NewRocket creates a new Rocket instance with default values.
func NewRocket(channel string) Rocket {
	return Rocket{
		Channel: channel,
	}
}

// GetKey returns the unique key for the Rocket, which is its channel.
// This key is used for storing and retrieving the rocket in a repository.
func (r Rocket) GetKey() string {
	return r.Channel
}

// UpdateState Contains the logic for each message type.
func (r *Rocket) UpdateState(messageType MessageType, messageData []byte) error {
	switch messageType {
	case RocketLaunched:
		var launchMsg LaunchedMessage
		if err := json.Unmarshal(messageData, &launchMsg); err != nil {
			return fmt.Errorf("unmarshal error - RocketLaunchedMessage: %w", err)
		}
		r.Type = launchMsg.Type
		r.Speed = launchMsg.LaunchSpeed
		r.Mission = launchMsg.Mission
		r.Exploded = false
		r.ExplosionReason = ""
		log.Printf("Rocket %s launched: Type=%s, Speed=%d, Mission=%s", r.Channel, r.Type, r.Speed, r.Mission)
	case RocketSpeedIncreased:
		var speedMsg SpeedChangedMessage
		if err := json.Unmarshal(messageData, &speedMsg); err != nil {
			return fmt.Errorf("unmarshal error - RocketSpeedIncreasedMessage: %w", err)
		}
		r.Speed += speedMsg.By
		log.Printf("Rocket %s speed increased by %d to %d", r.Channel, speedMsg.By, r.Speed)
	case RocketSpeedDecreased:
		var speedMsg SpeedChangedMessage
		if err := json.Unmarshal(messageData, &speedMsg); err != nil {
			return fmt.Errorf("unmarshal error - RocketSpeedDecreasedMessage: %w", err)
		}
		r.Speed -= speedMsg.By

		log.Printf("Rocket %s speed decreased by %d to %d", r.Channel, speedMsg.By, r.Speed)
	case RocketExploded:
		var explodedMsg RocketExplodedMessage
		if err := json.Unmarshal(messageData, &explodedMsg); err != nil {
			return fmt.Errorf("unmarshal error - RocketExplodedMessage: %w", err)
		}
		r.Exploded = true
		r.ExplosionReason = explodedMsg.Reason
		r.Speed = 0         // Speed becomes 0 upon explosion
		r.Mission = Aborted // Mission is aborted
		log.Printf("Rocket %s exploded! Reason: %s", r.Channel, r.ExplosionReason)
	case RocketMissionChanged:
		var missionMsg MissionChangedMessage
		if err := json.Unmarshal(messageData, &missionMsg); err != nil {
			return fmt.Errorf("unmarshal error - RocketMissionChangedMessage: %w", err)
		}
		r.Mission = missionMsg.NewMission
		log.Printf("Rocket %s mission changed to %s", r.Channel, r.Mission)
	default:
		log.Printf("Unknown message type received for %s: %s", r.Channel, messageType)
	}
	return nil
}
