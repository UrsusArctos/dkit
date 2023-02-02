package main

import (
	"fmt"
	"io/ioutil"
	"runtime"
	"time"

	mbot "github.com/UrsusArctos/dkit/umintgbot"
)

var tgb mbot.TGMinBotCore

func ActualHandler(msginfo mbot.TMessageInfo) {
	// Show received message
	fmt.Printf("%s [%d]: %s \n", msginfo.From.Username, msginfo.From.ID, msginfo.Text)
	// Send quoted reply
	sentmsg, err := tgb.SendMessage_AsReply(fmt.Sprintf("Hello, %s!", msginfo.From.Username), msginfo)
	if err != nil {
		fmt.Printf("%+v\n%+v\n", sentmsg, err)
	}
	/*
		// Send mp3 file
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
	token, _ := ioutil.ReadFile("token.txt")
	// Initialize bot
	tgb = mbot.NewInstance(string(token))
	fmt.Println("Started as @" + tgb.BotInfo.Result.Username)
	// Set message handler
	tgb.MSGHandler = ActualHandler
	// Set Buoy
	/*
		Buoy := buoy.TBuoyParams{
			MinimumSuccessTime:  3600,
			RestartDelay:        300,
			GeneralFailureCount: 10,
			DebugCallback:       DebugSayHandler}
	*/
	// Run message loop
	for {
		tgb.WatchForMessagesAsync()
		fmt.Print("[!]")
		for !tgb.LoadMessagesAsync() {
			runtime.Gosched()
			time.Sleep(1 * time.Millisecond)
		}
	}
	/*
		Buoy.KeepFloating(func() {

		})
	*/
}
