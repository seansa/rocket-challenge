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
