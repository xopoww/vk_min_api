package vk_min_api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
)

const (
	ApiAddr = "https://api.vk.com/"
)

func sendRequest(method string, values url.Values, token, version string)(*http.Response, error) {
	values.Set("access_token", token)
	values.Set("v", version)
	URL := fmt.Sprintf("%smethod/%s?%s", ApiAddr, method, values.Encode())
	return http.Get(URL)
}

/* Send a request to VK API.*/
func (bot * Bot) sendRequest(method string, values url.Values)(*http.Response, error) {
	return sendRequest(method, values, bot.token, bot.version)
}

/* Get user objects for the ids in the slice by requesting users.get VK API method */
func (bot * Bot) GetUsersByID(ids []int)([]User, error) {
	values := url.Values{}
	for _, id := range ids {
		values.Add("user_ids", string(id))
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
func (bot * Bot) GetUserByID(id int)(User, error){
	userSlice, err := bot.GetUsersByID([]int{id})
	if len(userSlice) > 0 {
		return userSlice[0], nil
	}
	return User{}, err
}

/* Send text message to user by id */
func (bot * Bot) SendMessage(to int, msg string)error {
	values := url.Values{}
	values.Set("user_id", fmt.Sprint(to))
	values.Set("random_id", fmt.Sprint(rand.Uint32()))
	values.Set("message", msg)

	bot.Logger.Debugf("Sending message to user (id = %d): %s", to, msg)

	resp, err := bot.sendRequest("messages.send", values)
	if err != nil {
		return err
	}
	var(
		body []byte
		respObj struct{
			Error string `json:"error"`
		}
	)
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.New(fmt.Sprintf("error reading body: %s", err))
	}
	bot.Logger.Debugf("Response body: %s", string(body))
	err = json.Unmarshal(body, &respObj)
	if err != nil {
		return errors.New(fmt.Sprintf("json: %s", err))
	}
	if respObj.Error == "" {
		return nil
	} else {
		return errors.New(fmt.Sprintf("API error: %s", respObj.Error))
	}
}