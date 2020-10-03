package vk_min_api

import "log"

type handlerPool struct {
	defaultText		func(*Message)
	commands		map[string]func(*Message)
	custom			[]customHandler
	callback		[]callbackHandler
}

type customHandler struct {
	condition		func(*Message)bool
	action			func(*Message)
}

type callbackHandler struct {
	condition		func(map[string]interface{})bool
	action			func(*MessageEvent)
}

/* Add the default Handler for messages
Added handler has lowest priority. Fires only if no other handler has. If this function is called
more than once, the default handler is taken from the last call (unrecommended).

If not set and no other handler has fired on a message, a warning is logged.*/
func (bot * Bot) HandleDefault(action func(*Message)) {
	bot.handlers.defaultText = action
}

/* Add the handler for the command. Has lower priority, than custom handlers, but higher priority,
than the default one.*/
func (bot * Bot) HandleOnCommand(command string, action func(*Message)) {
	bot.handlers.commands[command] = action
}

/* Add custom conditional handler.
Catches the message if condition(m) is true. Has the highest priority (i.e. if custom handler fires
on the message, neither command handlers not default handler will).

If several handlers fire on the same message, only the earliest one will run
(it is highly recommended to design more strict conditions for custom handlers).
 */
func (bot * Bot) HandleCustom(condition func(*Message)bool, action func(*Message)) {
	bot.handlers.custom = append(bot.handlers.custom, customHandler{condition, action})
}

/* Add callback handler
 */
func (bot * Bot) HandleCallback(condition func(map[string]interface{})bool, action func(*MessageEvent)) {
	bot.handlers.callback = append(bot.handlers.callback, callbackHandler{condition, action})
}

func (bot * Bot) handleNewMessage(m * Message) {

	// check custom handlers
	for _, hand := range bot.handlers.custom {
		if hand.condition(m) {
			hand.action(m)
			return
		}
	}
	// check command handlers
	if com := m.Command(); com != "" {
		if action, found := bot.handlers.commands[com]; found {
			action(m)
			return
		}
	}
	// apply default handler
	if bot.handlers.defaultText != nil {
		bot.handlers.defaultText(m)
		return
	}

	log.Printf("Unhandled message: %s\n", m.Text)
}

func (bot * Bot) handleMessageEvent(m * MessageEvent) {
	for _, hand := range bot.handlers.callback {
		if hand.condition(m.Payload) {
			hand.action(m)
			return
		}
	}

	log.Printf("Unhandeled message event: %+v", *m)
}