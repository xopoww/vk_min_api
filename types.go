package vk_min_api

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
)

type Message struct {
	ID				int						`json:"id"`
	Date			int						`json:"date"`

	FromID			int						`json:"from_id"`
	Text			string					`json:"text"`

	Payload			map[string]interface{}	`json:"payload"`
	Keyboard		Keyboard				`json:"keyboard"`

	//ReplyMessage	vkMessage				`json:"reply_message"` ???
}

var (
	reCom = regexp.MustCompile("/[^ ]+")
	reArg1 = regexp.MustCompile("/[^ ] .+")
	reArg2 = regexp.MustCompile(" .+")
)

func (m * Message) Command() string{
	match := reCom.Find([]byte(m.Text))
	if len(match) > 1 {
		return string(match[1:])
	}
	return ""
}

func (m * Message) CommandArg() string{
	match := reArg2.Find(reArg1.Find([]byte(m.Text)))
	if len(match) > 1 {
		return string(match[1:])
	}
	return ""
}

type Keyboard struct {
	OneTime			bool					`json:"one_time"`
	Buttons			[][]KeyboardButton		`json:"buttons"`
	Inline			bool					`json:"inline"`
}

func (k * Keyboard) MarshalJSON() ([]byte, error) {
	data := make(map[string]interface{})
	if k.Inline {
		data["inline"] = true
	} else {
		data["one_time"] = k.OneTime
	}
	data["buttons"] = k.Buttons
	return json.Marshal(data)
}

type KeyboardButton struct {
	Action			KeyboardAction			`json:"action"`
	Color			string					`json:"color"`
}

const (
	ColorBlue = "primary"
	ColorWhite = "secondary"
	ColorGreen = "positive"
	ColorRed = "negative"
)

type KeyboardAction struct {
	Type			string					`json:"type"`
	Label			string					`json:"label"`
	Payload			map[string]interface{}	`json:"payload"`
}

func NewCallbackButton(label, payload, color string) KeyboardButton {
	if color == "" {
		color = ColorWhite
	}
	return KeyboardButton{
		Action: KeyboardAction{
			Type:    "callback",
			Label:   label,
			Payload: map[string]interface{}{"data": payload},
		},
		Color:  color,
	}
}

type User struct {
	ID				int						`json:"id"`
	FirstName		string					`json:"first_name"`
	LastName		string					`json:"last_name"`
}

var ErrAPI = errors.New("API error")

func WrapApiErr(code int, msg string) error {
	return fmt.Errorf("%w %d: %s", ErrAPI, code, msg)
}