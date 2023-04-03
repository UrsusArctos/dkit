package main

const (
	ProjectName = "talkie"
	// Interfacing
	strExiting   = "Exiting"
	strStartedAs = "Started as @%s"
	strHint      = "Try writing meaningful sentences."
	strCleared   = "History cleared. The AI now does not remember previous conversation."
	strNoData    = "The model has returned neither an error nor data"
	// OpenAI
	PrefModel = "gpt-3.5-turbo"
	// SQL
	sqlCreateTable = `
	CREATE TABLE IF NOT EXISTS [chatlog] (
		[id] INTEGER  NOT NULL PRIMARY KEY AUTOINCREMENT,
		[tstamp] TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		[tgid] INTEGER  NOT NULL,
		[role] VARCHAR(10)  NOT NULL,
		[message] TEXT  NULL)`
	sqlRecordMessage   = `INSERT INTO chatlog (tgid, role, message) VALUES ($1,$2,$3)`
	sqlRetrieveHistory = `SELECT role,message FROM chatlog WHERE tgid=$1 ORDER BY tstamp ASC`
	sqlClearHistory    = `DELETE FROM chatlog WHERE tgid=$1`
	// Custom prefixes
	prefixImageGen  = "/imagen"
	prefixAwakening = "/awake"
)

type (
	TTalkieConfig struct {
		TGBotToken string `json:"TGBotToken"`
		OpenAIKey  string `json:"OpenAIKey"`
		ChatLogDB  string `json:"ChatLogDB"`
		AdminUID   int64  `json:"AdminUID"`
	}
)
