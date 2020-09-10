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

func sendRequest(method string, params map[string]interface{}, token, version string)(*http.Response, error) {
	vals := url.Values{}
	for key, values := range params {
		switch values.(type) {
		case []string:
			for _, value := range values.([]string) {
				vals.Add(key, value)
			}
		default:
			vals.Add(key, values.(string))
		}
	}
	vals.Set("access_token", token)
	vals.Set("v", version)
	URL := fmt.Sprintf("%smethod/%s?%s", ApiAddr, method, vals.Encode())
	return http.Get(URL)
}

/* Send a request to VK API.
if one of params values is a slice, it must be []string */
func (bot * Bot) sendRequest(method string, params map[string]interface{})(*http.Response, error) {
	return sendRequest(method, params, bot.token, bot.version)
}

/* Get user objects for the ids in the slice by requesting users.get VK API method */
func (bot * Bot) GetUsersByID(ids []int)([]User, error) {
	idStrs := make([]string, len(ids))
	for index, id := range ids {
		idStrs[index] = string(id)
	}
	params := map[string]interface{}{
		"user_ids": idStrs,
		"name_case": "nom",
	}
	resp, err := bot.sendRequest("users.get", params)
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
	params := map[string]interface{}{
		"user_id": to,
		"random_id": rand.Uint32(),
		"message": url.QueryEscape(msg),
	}
	resp, err := bot.sendRequest("messages.send", params)
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