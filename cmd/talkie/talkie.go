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
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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

// Ad-hoc blocklist. Needs to be reworked if the abuse problem persists
func (TB *TTalkieBot) IsBanned(visavis int64) bool {
	Banned := []int64{1468641954, 1493156936, 1699986538}
	var res bool = false
	for _, tgid := range Banned {
		res = res || (visavis == tgid)
	}
	return res
}

func (TB TTalkieBot) CheckSpecificError(prefix string, specerror error) {
	if specerror != nil {
		TB.Logger.LogEventError(fmt.Sprintf("%s:%v", prefix, specerror))
	}
}

func (TB *TTalkieBot) MessageHandler(msginfo kotobot.TMessage) {
	// Make sure the message is private and has text, ignore all other types
	if (msginfo.Chat.Type == "private") && (len(msginfo.Text) > 0) && (TB.IsBanned(msginfo.From.ID)) {
		TB.Logger.LogEventWarning(fmt.Sprintf("Banned TG ID [%d] attempts contact", msginfo.From.ID))
	}
	//
	if (msginfo.Chat.Type == "private") && (len(msginfo.Text) > 0) && (!TB.IsBanned(msginfo.From.ID)) {
		switch msginfo.Text {
		case "/start":
			{ // Give a hint
				TB.Logger.LogEventInfo(fmt.Sprintf("[%d] is greeting", msginfo.From.ID))
				_, senderr := TB.Bot.SendMessage(strHint, true, msginfo)
				TB.CheckSpecificError("Error sending greeting+hint message", senderr)
			}
		case "/models":
			{ // List available models reported by an API server
				TB.Logger.LogEventInfo(fmt.Sprintf("[%d] has requested list of models", msginfo.From.ID))
				var mtext strings.Builder
				for _, mdl := range TB.AIClient.Models.Data {
					mtext.WriteString(fmt.Sprintf("%s (%s)\n", mdl.ID, time.Unix(mdl.CreatedAt, 0).Format("Jan 2006")))
				}
				_, senderr := TB.Bot.SendMessage(mtext.String(), true, msginfo)
				TB.CheckSpecificError("Error sending models list", senderr)
			}
		case "/reset":
			{ // Clear locally recorded chat history
				TB.Logger.LogEventWarning(fmt.Sprintf("[%d] has requested history clearing", msginfo.From.ID))
				_, sqlerr := TB.S3DB.QuerySingle(sqlClearHistory, msginfo.From.ID)
				TB.CheckSpecificError("Database error", sqlerr)
				_, senderr := TB.Bot.SendMessage(strCleared, true, msginfo)
				TB.CheckSpecificError("Error sending clearing reply", senderr)
			}
		default:
			{ // Check if this is a custom admin command
				if strings.HasPrefix(msginfo.Text, prefixAwakening) && (msginfo.From.ID == TB.Config.AdminUID) {
					// admin-only command of awakening
					// extract UID
					spls := strings.Split(msginfo.Text, " ")
					fmt.Printf("%+v\n", spls)
					var ruid int64
					fmt.Sscanf(spls[1], "%d", &ruid)
					OMsg := strings.Join(spls[2:], " ")
					// Log attempt
					TB.Logger.LogEventInfo(fmt.Sprintf("[%d]-awake to %d: %s", msginfo.From.ID, ruid, OMsg))
					// Send message
					rRefMsg := kotobot.TMessage{Chat: &tgbotapi.Chat{ID: ruid}}
					_, awerr := TB.Bot.SendMessage(OMsg, false, rRefMsg)
					TB.CheckSpecificError("Awake msg error: ", awerr)
					break
				}
				// imagen
				if strings.HasPrefix(msginfo.Text, prefixImageGen) && (msginfo.From.ID == TB.Config.AdminUID) {
					// admin-only command of imagen
					// extract prompt
					_, prompt, _ := strings.Cut(msginfo.Text, prefixImageGen)
					prompt = strings.TrimSpace(prompt)
					// Log attempt
					TB.Logger.LogEventInfo(fmt.Sprintf("[%d]-%s: %s", msginfo.From.ID, prefixImageGen, prompt))
					// Generate
					gi, gierr := TB.AIClient.GetGeneratedImage(prompt, 1)
					TB.CheckSpecificError("Error generating image", gierr)
					if gierr == nil {
						// && ()
						if len(gi.Data) > 0 {
							_, senderr := TB.Bot.SendMessage(gi.Data[0].URL, true, msginfo)
							TB.CheckSpecificError("Error sending URL of image", senderr)
						} else {
							TB.Logger.LogEventError(strNoData)
							_, senderr := TB.Bot.SendMessage(strNoData, true, msginfo)
							TB.CheckSpecificError("Error sending error message", senderr)
						}
					}
					//
					break
				}
				// default: regular chat message
				// Log message
				TB.Logger.LogEventInfo(fmt.Sprintf("[%d]: %s", msginfo.From.ID, msginfo.Text))
				// Record this message in history
				_, sqlerr := TB.S3DB.QuerySingle(sqlRecordMessage, msginfo.From.ID, openai.ChatRoleUser, msginfo.Text)
				TB.CheckSpecificError("Database error", sqlerr)
				// Handle message
				TB.handlePrivateMessage(msginfo)
			}
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
			TB.CheckSpecificError("Error sending reply message", senderr)
			// 4. Record message
			_, recerr := TB.S3DB.QuerySingle(sqlRecordMessage, msginfo.From.ID, openai.ChatRoleAssistant, Reply.Content)
			TB.CheckSpecificError("Error recording reply message", recerr)
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
