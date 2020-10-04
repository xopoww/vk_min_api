package vk_min_api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
)

const (
	ApiAddr = "https://api.vk.com/"
)

// Send a request to VK API.
//
// If API response contains an error, wrapped ErrAPI will be returned. Otherwise, the "response" field
// of response body will be unmarshalled into dst.
func (bot * Bot) sendRequest(method string, values url.Values, dst interface{}) error  {
	values.Set("access_token", bot.token)
	values.Set("v", bot.version)
	URL := fmt.Sprintf("%smethod/%s?%s", ApiAddr, method, values.Encode())
	if bot.verbose {
		log.Printf("Sending a request: %s", URL)
	}
	resp, err :=  http.Get(URL)
	if err != nil {
		return fmt.Errorf("http GET: %w", err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read body: %w", err)
	}
	respObj := struct {
		Error struct {
			Code int `json:"error_code"`
			Msg string `json:"error_msg"`
		} `json:"error"`
		Response interface{} `json:"response"`
	}{
		Response: dst,
	}
	err = json.Unmarshal(body, &respObj)
	if err != nil {
		return fmt.Errorf("unmarshal body: %w", err)
	}
	if respObj.Error.Msg != "" {
		return WrapApiErr(respObj.Error.Code, respObj.Error.Msg)
	}
	return nil
}

/* Get user objects for the ids in the slice by requesting users.get VK API method */
func (bot * Bot) GetUsersByID(ids []int)([]User, error) {
	values := url.Values{}
	for _, id := range ids {
		values.Add("user_ids", fmt.Sprint(id))
	}

	var users []User
	err := bot.sendRequest("users.get", values, &users)
	return users, err
}

/* Wrapper for GetUsersByID for only one user
calls GetUsersByID inside of it */
func (bot * Bot) GetUserByID(id int)(User, error) {
	userSlice, err := bot.GetUsersByID([]int{id})
	if len(userSlice) > 0 {
		return userSlice[0], nil
	}
	return User{}, err
}

/* Send text message to user by id
If keyboard is not nil, it is attached to the message.*/
func (bot * Bot) SendMessage(to int, msg string, keyboard *Keyboard) (int, error) {
	values := url.Values{}
	values.Set("user_id", fmt.Sprint(to))
	values.Set("random_id", fmt.Sprint(rand.Uint32()))
	values.Set("message", msg)
	if keyboard != nil {
		keyboardData, err := json.Marshal(keyboard)
		if err != nil {
			return 0, fmt.Errorf("json: %w", err)
		}
		values.Set("keyboard", string(keyboardData))
	}

	var id int
	err := bot.sendRequest("messages.send", values, &id)
	return id, err
}

func (bot * Bot) GetMessagesByID(ids []int)([]Message, error) {
	values := url.Values{}
	for _, id := range ids {
		values.Add("message_ids", fmt.Sprint(id))
	}
	//values.Set("group_id", fmt.Sprint(bot.groupID))
	var messages []Message
	err := bot.sendRequest("messages.getByID", values, &messages)
	return messages, err
}

func(bot * Bot) GetMessageByID(id int)(Message, error) {
	messages, err := bot.GetMessagesByID([]int{id})
	if len(messages) > 0 {
		return messages[0], nil
	}
	return Message{}, err
}

func(bot * Bot) GetMessagesByConversationID(peerID int, ids []int)([]Message, error) {
	values := url.Values{}
	for _, id := range ids {
		values.Add("conversation_message_ids", fmt.Sprint(id))
	}
	values.Add("peer_id", fmt.Sprint(peerID))
	var messages []Message
	err := bot.sendRequest("messages.getByConversationMessageId", values, &messages)
	return messages, err
}

func(bot * Bot) GetMessageByConversationID(peerID, id int)(Message, error) {
	messages, err := bot.GetMessagesByConversationID(peerID, []int{id})
	if len(messages) > 0 {
		return messages[0], nil
	}
	return Message{}, err
}

func (bot * Bot) EditMessage(peerID, messageID int, msg string, keyboard *Keyboard) error {
	values := url.Values{}
	values.Set("peer_id", fmt.Sprint(peerID))
	values.Set("message_id", fmt.Sprint(messageID))
	values.Set("message", msg)
	if keyboard != nil {
		keyboardData, err := json.Marshal(keyboard)
		if err != nil {
			return fmt.Errorf("json: %w", err)
		}
		values.Set("keyboard", string(keyboardData))
	}

	return bot.sendRequest("messages.edit", values, nil)
}

// 	Send an answer to a message event
// Currently supports only "show_snackbar" answer type which will display an answer as a text to user.
// It will do nothing if answer is "".
//
// Note, that every message event must be answered within 1 minute.
func (bot * Bot) SendMessageEventAnswer(event *MessageEvent, answer string) error {
	values := url.Values{}
	values.Set("event_id", event.EventID)
	values.Set("user_id", fmt.Sprint(event.UserID))
	values.Set("peer_id", fmt.Sprint(event.PeerID))
	values.Set("event_data", fmt.Sprintf(`{"type":"show_snackbar","text":"%s"}`, answer))

	return bot.sendRequest("messages.sendMessageEventAnswer", values, nil)
}