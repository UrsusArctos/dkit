package main

import (
	"fmt"
	"os"

	"github.com/UrsusArctos/dkit/pkg/openai"
)

const (
	PrefModel = "gpt-3.5-turbo"
)

func main() {
	OpenAIKey, _ := os.ReadFile(".creds/token-openai.txt")
	oac, err := openai.NewInstance(string(OpenAIKey))
	if err == nil {
		if oac.SelectModel(PrefModel) {
			chat := openai.NewChat()
			chat.Say("Suggest the name for my new dog")
			cc, _ := oac.GetChatCompletion(chat, 2)
			chat.PickAnswer(cc, 1)
			chat.Say("Thank you!")
			cc, _ = oac.GetChatCompletion(chat, 2)
			chat.PickAnswer(cc)
			fmt.Printf("CHAT: %+v\n", chat)
		}
	}
}
