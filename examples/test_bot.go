package main

import (
	"fmt"
	vk "github.com/xopoww/vk_min_api"
	"log"
	"os"
	"strconv"
)

func main() {
	keyboard := vk.Keyboard{
		Inline: true,
		Buttons: [][]vk.KeyboardButton{
			{
				vk.NewCommandButton("Foo", "but", 1, vk.ColorGreen),
				vk.NewCommandButton("Bar", "but", 2, vk.ColorRed),
			},
			{
				vk.NewDataButton("Add counter", "add", vk.ColorBlue),
			},
		},
	}

	bot, err := vk.NewBot(vk.Properties{
		Token: os.Getenv("VK_TOKEN"),
		Version: "5.95",
		Secret: "testing",
		VerboseLogging: true,
	},false)
	if err != nil {
		panic(err)
	}

	bot.HandleDefault(
		func(m *vk.Message){
			_, err := bot.SendMessage(m.FromID, "1", &keyboard)
			if err != nil {
				log.Printf("Send message: %s", err)
			}
		})

	bot.HandleOnCommand("but", func(m * vk.MessageEvent){
		arg, found := m.Payload["arg"]
		argInt, ok := arg.(int)
		var answer string
		switch {
		case !found || !ok:
			answer = "Bad payload."
		case argInt == 1:
			answer = "Foo!"
		case argInt == 2:
			answer = "Bar!"
		default:
			answer = "Unknown button."
		}
		err := bot.SendMessageEventAnswer(m, answer)
		if err != nil {
			log.Printf("Send message event answer: %s", err)
		}
	})

	bot.HandleCallback(
		func(data interface{}) bool {
			if dataStr, ok := data.(string); ok {
				return dataStr == "add"
			}
			return false
		},
		func(m *vk.MessageEvent){
			err := bot.SendMessageEventAnswer(m, "")
			if err != nil {
				log.Printf("Send message event answer: %s", err)
				return
			}

			msg, err := bot.GetMessageByConversationID(m.PeerID, m.MessageID)
			if err != nil {
				log.Printf("Cannot get message: %s", err)
				return
			}
			count, err := strconv.Atoi(msg.Text)
			if err != nil {
				log.Printf("Atoi: %s", err)
				return
			}
			err = bot.EditMessage(m.PeerID, msg.ID, fmt.Sprint(count+1), &keyboard)
			if err != nil {
				log.Printf("Edit message: %s", err)
			}
		})

	bot.StartWithServer("", "/vk")
}
