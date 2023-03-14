package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/UrsusArctos/dkit/pkg/daemonizer"
	"github.com/UrsusArctos/dkit/pkg/kotobot"
	"github.com/UrsusArctos/dkit/pkg/logmeow"
	"github.com/UrsusArctos/dkit/pkg/openai"
	"github.com/UrsusArctos/dkit/pkg/sqlite"
)

type (
	TTalkieBot struct {
		Config      TTalkieConfig
		LinuxDaemon daemonizer.TLinuxDaemon
		Logger      logmeow.TLogMeow
		Bot         kotobot.TKotoBot
		S3DB        sqlite.TSQLite3DB
		AIClient    openai.TOpenAPIClient
	}
)

func (TB *TTalkieBot) BotInit() (err error) {
	// Check if config file exists
	_, cferr := os.Stat(TB.LinuxDaemon.ConfFile)
	if cferr == nil {
		cfjson, _ := os.ReadFile(TB.LinuxDaemon.ConfFile)
		unerr := json.Unmarshal(cfjson, &TB.Config)
		if unerr == nil {
			// Config loaded successfully
			// 1. Initialize chatlog database
			var dberr error
			TB.S3DB, dberr = sqlite.ConnectToDB(TB.Config.ChatLogDB)
			if dberr != nil {
				return dberr
			}

			// 2. Make sure chatlog table exists
			_, dberr = TB.S3DB.QuerySingle(sqlCreateTable)
			if dberr != nil {
				return dberr
			}
			TB.Logger.LogEventInfo("Database OK")

			// 3. Initialize OpenAPI client
			var aierr error
			TB.AIClient, aierr = openai.NewInstance(TB.Config.OpenAIKey)
			if aierr != nil {
				return aierr
			}
			if TB.AIClient.SelectModel(PrefModel) {
				TB.Logger.LogEventInfo("OpenAI client OK")
			} else {
				return fmt.Errorf("unable to select completion model")
			}

			// 4. Initialize Telegram bot
			var boterr error
			TB.Bot, boterr = kotobot.NewInstance(TB.Config.TGBotToken)
			if boterr != nil {
				return boterr
			}
			TB.Bot.MessageHandler = TB.MessageHandler
			TB.Bot.ParseMode = kotobot.PMPlainText
			TB.Bot.Updates_StartWatch()

			// X. Initialization done
			TB.Logger.LogEventInfo(fmt.Sprintf(strStartedAs, TB.Bot.BotInfo.UserName))
		}
		return unerr
	}
	return cferr
}

func (TB *TTalkieBot) BotClose() (err error) {
	TB.S3DB.Close()
	return nil
}

func (TB *TTalkieBot) BotMain() (err error) {
	if TB.Bot.Updates_ProcessAll() {
		TB.Bot.Updates_StartWatch()
	} else {
		time.Sleep(100 * time.Millisecond)
	}
	return nil
}

func (TB *TTalkieBot) MessageHandler(msginfo kotobot.TMessage) {
	// Make sure the message is private and has text, ignore all other types
	if (msginfo.Chat.Type == "private") && (len(msginfo.Text) > 0) {
		var senderr error
		switch msginfo.Text {
		case "/start":
			{ // Give a hint
				_, senderr = TB.Bot.SendMessage(strHint, true, msginfo)
			}
		case "/models":
			{ // List available models reported by an API server
				var mtext strings.Builder
				for _, mdl := range TB.AIClient.Models.Data {
					mtext.WriteString(fmt.Sprintf("%s (%s)\n", mdl.ID, time.Unix(mdl.CreatedAt, 0).Format("Jan 2006")))
				}
				_, senderr = TB.Bot.SendMessage(mtext.String(), true, msginfo)
			}
		case "/reset":
			{ // Clear locally recorded chat history
				_, sqlerr := TB.S3DB.QuerySingle(sqlClearHistory, msginfo.From.ID)
				if sqlerr != nil {
					TB.Logger.LogEventError(fmt.Sprintf("Database error: %+v", sqlerr))
				}
				_, senderr = TB.Bot.SendMessage(strCleared, true, msginfo)
			}
		default:
			{ // Record this message in history
				_, sqlerr := TB.S3DB.QuerySingle(sqlRecordMessage, msginfo.From.ID, openai.ChatRoleUser, msginfo.Text)
				if sqlerr != nil {
					TB.Logger.LogEventError(fmt.Sprintf("Error recording message: %+v", sqlerr))
				}
				// Log message
				TB.Logger.LogEventInfo(fmt.Sprintf("[%d]: %s", msginfo.From.ID, msginfo.Text))
				// Handle message
				TB.handlePrivateMessage(msginfo)
			}
		}
		// React to error
		if senderr != nil {
			TB.Logger.LogEventError(fmt.Sprintf("Error sending message: %+v", senderr))
		}
	}
}

func (TB *TTalkieBot) handlePrivateMessage(msginfo kotobot.TMessage) {
	// 1. Collect all chat history
	chatCurrent := openai.NewChat()
	defer chatCurrent.Clear()
	chist, cherr := TB.S3DB.QueryData(sqlRetrieveHistory, msginfo.From.ID)
	if cherr == nil {
		defer chist.Close()
		for {
			datarow := chist.UnloadNextRow()
			if datarow == nil {
				break
			}
			chatCurrent.RecordMessage(datarow["role"], datarow["message"])
		}
		// 2. Get Completion from OpenAI
		ChatCom, ccerr := TB.AIClient.GetChatCompletion(chatCurrent, 1)
		if ccerr == nil {
			Reply := chatCurrent.PickAnswer(ChatCom, 0)
			// 3. Send reply
			_, senderr := TB.Bot.SendMessage(Reply.Content, true, msginfo)
			if senderr != nil {
				TB.Logger.LogEventError(fmt.Sprintf("Error sending reply message: %+v", senderr))
			}
			// 4. Record message
			_, recerr := TB.S3DB.QuerySingle(sqlRecordMessage, msginfo.From.ID, openai.ChatRoleAssistant, Reply.Content)
			if recerr != nil {
				TB.Logger.LogEventError(fmt.Sprintf("Error recording reply message: %+v", recerr))
			}
			// 5. Report reply
			TB.Logger.LogEventInfo(fmt.Sprintf("[openai]: %s", Reply.Content))
			return
		}
		TB.Logger.LogEventError(fmt.Sprintf("Error getting completion: %+v", ccerr))
		return
	}
	TB.Logger.LogEventError(fmt.Sprintf("Error retrieving specific history: %+v", cherr))
}

func main() {
	// Init daemon
	talkie := TTalkieBot{LinuxDaemon: daemonizer.NewLinuxDaemon(ProjectName)}
	defer talkie.LinuxDaemon.Close()
	talkie.LinuxDaemon.FuncInit = talkie.BotInit
	talkie.LinuxDaemon.FuncClose = talkie.BotClose
	talkie.LinuxDaemon.FuncMain = talkie.BotMain
	// Init logger
	var enfac uint8 = logmeow.FacFile
	if talkie.LinuxDaemon.Foreground {
		enfac |= logmeow.FacConsole
	}
	talkie.Logger = logmeow.NewLogMeow(ProjectName, enfac, talkie.LinuxDaemon.LogPath)
	defer talkie.Logger.Close()
	// Run daemon
	derror := talkie.LinuxDaemon.Run()
	if derror != nil {
		talkie.Logger.LogEventError(fmt.Sprintf("%s: %v", strExiting, derror))
	} else {
		talkie.Logger.LogEventInfo(strExiting)
	}
}
