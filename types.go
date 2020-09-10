package vk_min_api

import "regexp"

type Message struct {
	ID				int						`json:"id"`
	Date			int						`json:"date"`

	FromID			int						`json:"from_id"`
	Text			string					`json:"text"`

	Payload			string					`json:"payload"`
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

type KeyboardButton struct {
	Action			KeyboardAction			`json:"action"`
	Color			string					`json:"color"`
}

type KeyboardAction struct {
	Type			string					`json:"type"`
	Label			string					`json:"label"`
	Payload			string					`json:"payload"`
}

type User struct {
	ID				int						`json:"id"`
	FirstName		string					`json:"first_name"`
	LastName		string					`json:"last_name"`
}
