package vk_min_api

type Handler struct {
	Condition		func(*Message)bool
	Action			func(*Message)
	Priority		int
}

/* Add the default Handler for text messages
Added handler has lowest priority. It catches not only plain text,
but also unhandled commands.*/
func (bot * Bot) HandleOnText(action func(*Message)) {
	bot.handlers = append(bot.handlers, Handler{func(m * Message)bool{ return true }, action, 0})
}

/* Add the handler for the command */
func (bot * Bot) HandleOnCommand(command string, action func(*Message)) {
	hand := Handler{
		Condition: func(m *Message) bool {
			return m.Command() == command
		},
		Action:   action,
		Priority: 5,
	}
	bot.handlers = append(bot.handlers, hand)
}

func (bot * Bot) handleNewMessage(m * Message) {
	bot.Logger.Debugf("Handling new message (choosing from %d handlers): %s", len(bot.handlers), m.Text)

	handMatch := &Handler{Priority: -1}
	for _, hand := range bot.handlers {
		if hand.Priority > handMatch.Priority && hand.Condition(m) {
			handMatch = &hand
		}
	}
	if handMatch.Priority == -1 {
		bot.Logger.Debug("No suitable handlers found!")
		return
	}
	handMatch.Action(m)
}