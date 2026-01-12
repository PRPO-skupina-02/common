package messaging

import (
	"encoding/json"
	"time"
)

type EmailMessage struct {
	To string `json:"to" validate:"required,email"`

	Template string `json:"template" validate:"required"`

	TemplateData map[string]interface{} `json:"template_data,omitempty"`

	CreatedAt time.Time `json:"created_at"`
}

func NewEmailMessage(to, template string, data map[string]interface{}) *EmailMessage {
	return &EmailMessage{
		To:           to,
		Template:     template,
		TemplateData: data,
		CreatedAt:    time.Now(),
	}
}

func (e *EmailMessage) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

func FromJSON(data []byte) (*EmailMessage, error) {
	var msg EmailMessage
	err := json.Unmarshal(data, &msg)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}
