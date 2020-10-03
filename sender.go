package vk_min_api

import (
	"encoding/json"
	"errors"
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

/* Send a request to VK API.*/
func (bot * Bot) sendRequest(method string, values url.Values)(*http.Response, error) {
	values.Set("access_token", bot.token)
	values.Set("v", bot.version)
	URL := fmt.Sprintf("%smethod/%s?%s", ApiAddr, method, values.Encode())
	if bot.verbose {
		log.Printf("Sending a request: %s", URL)
	}
	return http.Get(URL)
}

/* Get user objects for the ids in the slice by requesting users.get VK API method */
func (bot * Bot) GetUsersByID(ids []int)([]User, error) {
	values := url.Values{}
	for _, id := range ids {
		values.Add("user_ids", fmt.Sprint(id))
	}
	resp, err := bot.sendRequest("users.get", values)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("API error: %s", err))
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error reading body: %s", err))
	}
	respObj := struct{Response []User `json:"response"`}{}
	err = json.Unmarshal(body, &respObj)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("json: %s", err))
	}
	return respObj.Response, nil
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
func (bot * Bot) SendMessage(to int, msg string, keyboard *Keyboard) error {
	values := url.Values{}
	values.Set("user_id", fmt.Sprint(to))
	values.Set("random_id", fmt.Sprint(rand.Uint32()))
	values.Set("message", msg)
	if keyboard != nil {
		keyboardData, err := json.Marshal(keyboard)
		if err != nil {
			return fmt.Errorf("json: %w", err)
		}
		values.Set("keyboard", string(keyboardData))
	}

	resp, err := bot.sendRequest("messages.send", values)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}

	return parseApiResponse(resp)
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

	resp, err := bot.sendRequest("messages.sendMessageEventAnswer", values)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}

	return parseApiResponse(resp)
}

func parseApiResponse(resp * http.Response) error {
	var respObj struct {
		Error struct {
				Code int `json:"error_code"`
				Msg string `json:"error_msg"`
			} `json:"error"`
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read body: %w", err)
	}
	err = json.Unmarshal(body, &respObj)
	if err != nil {
		return fmt.Errorf("unmarshal body: %w", err)
	}
	if respObj.Error.Msg == "" {
		return nil
	} else {
		return WrapApiErr(respObj.Error.Code, respObj.Error.Msg)
	}
}