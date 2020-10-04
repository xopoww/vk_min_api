package main

import (
	vk "github.com/xopoww/vk_min_api"
	"log"
	"os"
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

	var messageID int

	bot.HandleDefault(
		func(m *vk.Message){
			mid, err := bot.SendMessage(m.FromID, "1", &keyboard)
			if err != nil {
				log.Printf("Send message: %s", err)
			}
			messageID = mid
		})

	bot.HandleOnCommand("but", func(m * vk.MessageEvent){
		arg, found := m.Payload["arg"]
		argFloat, ok := arg.(float64)
		argInt := int(argFloat)
		var answer string
		switch {
		case !found:
			answer = "Arg not found."
		case !ok:
			answer = "Arg not int."
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
			err := bot.SendMessageEventAnswer(m, "Incremented.")
			if err != nil {
				log.Printf("Send message event answer: %s", err)
				return
			}

			msg, err := bot.GetMessageByID(messageID)
			if err != nil {
				log.Printf("Cannot get message: %s", err)
				return
			}
			err = bot.EditMessage(m.PeerID, msg.ID, msg.Text + " a", &keyboard)
			if err != nil {
				log.Printf("Edit message: %s", err)
			}
		})

	log.Println("Stating the bot...")
	bot.StartWithServer("", "/vk")
}
