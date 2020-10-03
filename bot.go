package vk_min_api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
)

type Bot struct {
	token				string
	version				string
	callbackConfirmed	bool
	callbackConfig		*CallbackConfig
	secret				string
	//Logger

	requestsChan		chan []byte
	waitGroup			sync.WaitGroup

	handlers			handlerPool

	verbose				bool
}

type CallbackConfig struct {
	ReqProcessors	int
	ReqChanSize		int
	Confirmation	*CallbackConfirmation
}
const (
	defaultReqProcessors = 5
	defaultReqChanSize = 25
)

type CallbackConfirmation struct {
	GroupID		int
	Response	string
}

type Properties struct {
	Token string
	Version string
	Secret string
	CallbackProps *CallbackConfig
	VerboseLogging bool
}

func NewBot(properties Properties, listenForConfirmation bool)(*Bot, error) {
	if properties.CallbackProps == nil {
		properties.CallbackProps = &CallbackConfig{
			ReqProcessors: defaultReqProcessors,
			ReqChanSize: defaultReqChanSize,
		}
	} else {
		if properties.CallbackProps.ReqProcessors <= 0 {
			return nil, errors.New("ReqProcessors must be positive")
		}
		if properties.CallbackProps.ReqChanSize < 0 {
			return nil, errors.New("ReqChanSize must be non-negative")
		}
		if properties.CallbackProps.Confirmation == nil && listenForConfirmation {
			return nil, errors.New("CallbackConfirmation object is required if listenForConfirmation is true")
		}
	}

	bot := Bot{
		token: properties.Token,
		version: properties.Version,
		callbackConfirmed: !listenForConfirmation,
		callbackConfig: properties.CallbackProps,
		secret: properties.Secret,
		//Logger: logger,
		requestsChan: make(chan []byte, properties.CallbackProps.ReqChanSize),
		handlers: handlerPool{
			commands:    map[string]func(*Message){},
		},
		verbose: properties.VerboseLogging,
	}

	return &bot, nil
}

// starts the bot request processors
// blocking, use bot.Stop() to unblock
// NOTE: for CallbackAPI requests to be received (and processed) one must
// start an http server with bot.HTTPHandler()
func (bot * Bot) Start() {
	for i := 0; i < bot.callbackConfig.ReqProcessors; i++ {
		go func() {
			req := <- bot.requestsChan
			err := bot.processRequest(req)
			if err != nil {
				log.Printf("Error processing request: %s\n", err)
			}
		}()
	}
	bot.waitGroup.Add(1)
	bot.waitGroup.Wait()
}

// same as bot.Start(), but also starts http.ListenAndServe
func (bot * Bot) StartWithServer(addr, endpoint string) {
	http.HandleFunc(endpoint, bot.HTTPHandler())
	go func() {
		err := http.ListenAndServe(addr, nil)
		bot.Stop()
		log.Fatalf("HTTP Server failed: %s", err)
	}()
	bot.Start()
}

// stops the bot
func (bot * Bot) Stop() {
	bot.waitGroup.Done()
}

// returns http handler which will listen for CallbackAPI requests
// example usage:
// http.HandleFunc("/vk", bot.HTTPHandler())
func (bot * Bot) HTTPHandler() func(http.ResponseWriter, *http.Request){
	return func(w http.ResponseWriter, r *http.Request){
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Printf("Error reading request body: %s\n", err)
			return
		}

		if !bot.callbackConfirmed {
			var confirmation struct {
				Type	string	`json:"type"`
				GroupID	int		`json:"group_id"`
			}
			err = json.Unmarshal(body, &confirmation)
			if err != nil {
				log.Printf("json error: %s\n", err)
				return
			}
			if confirmation.Type == "confirmation" && confirmation.GroupID == bot.callbackConfig.Confirmation.GroupID {
				_, err = fmt.Fprint(w, bot.callbackConfig.Confirmation.Response)
				if err != nil {
					log.Printf("Error writing a response to confirmation request: %s\n", err)
				} else {
					bot.callbackConfirmed = true
				}
				return
			}
		}

		if bot.verbose {
			log.Printf("Got a request. Body: %s", string(body))
		}

		bot.requestsChan <- body

		_, err = fmt.Fprint(w, "ok")
		if err != nil {
			log.Printf("Error writing a response: %s\n", err)
		}
	}
}

// request processing

// processes a request according to its type
func (bot * Bot) processRequest(body []byte)error {
	reqType, err := bot.getRequestType(body)
	if err != nil {
		return errors.New(fmt.Sprintf("get request type: %s", err))
	}

	switch reqType {
	case "message_new":
		var obj struct{Object Message `json:"object"`}
		err = json.Unmarshal(body, &obj)
		if err != nil {
			return fmt.Errorf("json: %w", err)
		}
		bot.handleNewMessage(&obj.Object)
		return nil
	case "bad_secret":
		log.Println("Got a request with a wrong secret")
	default:
		log.Printf("Unsupported request type: %s\n", reqType)
	}

	return nil
}

// reads request type and returns it
// also checks for API secret and
// returns "bad_secret" in case of mismatch
func (bot * Bot) getRequestType(body []byte)(string, error) {
	var ar struct{
		Type			string					`json:"type"`
		Secret			string					`json:"secret"`
	}
	err := json.Unmarshal(body, &ar)
	if err != nil {
		return "", fmt.Errorf("json: %w", err)
	}
	if ar.Secret == bot.secret {
		return ar.Type, nil
	} else {
		return "bad_secret", nil
	}
}