package main

import (
	"fmt"
	"os"
	"time"

	"github.com/UrsusArctos/dkit/pkg/kotobot"
)

var tgb kotobot.TKotoBot

func ActualHandler(msginfo kotobot.TMessage) {
	// Show received message
	fmt.Printf("%s [%d]: %s \n", msginfo.From.UserName, msginfo.From.ID, msginfo.Text)
	// Send quoted reply
	sentmsg, err := tgb.SendMessage(fmt.Sprintf("Hello, %s!", msginfo.From.UserName), true, msginfo)
	if err != nil {
		fmt.Printf("%+v\n%+v\n", sentmsg, err)
	}
	// Send picture as file
	sentpicmsg, err := tgb.SendDocument(msginfo.From.ID, "cat.jpg", fmt.Sprintf("Downloaded using @%s", tgb.BotInfo.UserName))
	if err != nil {
		fmt.Printf("%+v\n%+v\n", sentpicmsg, err)
	}
}

func main() {
	// Read Bot API token from file
	token, _ := os.ReadFile("token.txt")
	// Initialize bot
	tgb, _ = kotobot.NewInstance(string(token))
	fmt.Println("Started as @" + tgb.BotInfo.UserName)
	// Set message handler
	tgb.MessageHandler = ActualHandler
	// Run message loop
	tgb.Updates_StartWatch()
	for {
		if tgb.Updates_ProcessAll() {
			fmt.Print("#")
			tgb.Updates_StartWatch()
		} else {
			fmt.Print(".")
			time.Sleep(500 * time.Millisecond)
		}
	}
}
