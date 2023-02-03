package main

import (
	"fmt"
	"os"
	"time"

	"github.com/UrsusArctos/dkit/pkg/umintgbot"
)

var tgb umintgbot.TGMinBotCore

func ActualHandler(msginfo umintgbot.TMessageInfo) {
	// Show received message
	fmt.Printf("%s [%d]: %s \n", msginfo.From.Username, msginfo.From.ID, msginfo.Text)
	// Send quoted reply
	sentmsg, err := tgb.SendMessage_AsReply(fmt.Sprintf("Hello, %s!", msginfo.From.Username), msginfo)
	if err != nil {
		fmt.Printf("%+v\n%+v\n", sentmsg, err)
	}
	// This is how to send and MP3 file
	/*
		afile := mbot.AttachedFileData{LocalFile: "sample.mp3",
			Caption: "Downloaded using @" + tgb.BotInfo.Result.Username, Performer: "Demo", Title: "Sample Sound",
		}
		sentaudiomsg, err := tgb.SendMessage_Audio(afile, msginfo.From.ID)
		if err != nil {
			fmt.Printf("%+v\n%+v\n", sentaudiomsg, err)
		}
	*/
}

func DebugSayHandler(message string) {
	fmt.Println(message)
}

func main() {
	// Read Bot API token from file
	token, _ := os.ReadFile("token.txt")
	// Initialize bot
	tgb = umintgbot.NewInstance(string(token))
	fmt.Println("Started as @" + tgb.BotInfo.Result.Username)
	// Set message handler
	tgb.MSGHandler = ActualHandler
	// Run message loop
	for {
		tgb.WatchForMessagesAsync()
		fmt.Print("[!]")
		for tgb.LoadMessagesAsync() {
			// probably replace this sleep with runtime.Gosched() in high-load applications
			time.Sleep(1 * time.Millisecond)
		}
	}
}
