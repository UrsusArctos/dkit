package kotbot

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/UrsusArctos/dkit/pkg/dexternal"

	tgtypes "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	// URLs and methods
	urlAPI       = "https://api.telegram.org/bot%s/%s"
	httpMIMEType = "application/json"
	uriGetMe     = "getMe"
)

type (
	TKotBot struct {
		// Public fields
		APIToken string // Telegram Bot API token
		BotInfo  tgtypes.User
		// Private fields
		dext *dexternal.TDexternal
	}
)

// NewInstance creates a new instance of the KotBot
func NewInstance(token string) (*TKotBot, error) {
	resp, err := http.Post(TKotBot{APIToken: token}.formatURL(uriGetMe), httpMIMEType, nil)
	if (err == nil) && (resp.StatusCode == http.StatusOK) {
		defer resp.Body.Close()
		respBytes, err := io.ReadAll(resp.Body)
		if err == nil {
			var apires tgtypes.APIResponse
			err = json.Unmarshal(respBytes, &apires)
			if err == nil {
				kb := &TKotBot{APIToken: token}
				err = json.Unmarshal(apires.Result, &kb.BotInfo)
				if err == nil {
					return kb, nil
				}
			}
		}
	}
	//
	return nil, err
}

func (KB TKotBot) formatURL(apiMethod string) string {
	return fmt.Sprintf(urlAPI, KB.APIToken, apiMethod)
}
