package model

// Rocket represents the current state of a rocket.
type Rocket struct {
	Channel         string `json:"channel"`
	Type            string `json:"type,omitempty"`
	Speed           int    `json:"speed"`
	Mission         string `json:"mission,omitempty"`
	Exploded        bool   `json:"exploded"`
	ExplosionReason string `json:"explosionReason,omitempty"`
}

// NewRocket creates a new Rocket instance with default values.
func NewRocket(channel string) *Rocket {
	return &Rocket{
		Channel: channel,
	}
}

// GetKey returns the unique key for the Rocket, which is its channel.
// This key is used for storing and retrieving the rocket in a repository.
func (r Rocket) GetKey() string {
	return r.Channel
}
