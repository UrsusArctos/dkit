package upsdclient

const (
	// action logic
	LoadOffTimeout = 15
	// bracket keywords
	kwBEGIN = "BEGIN"
	kwEND   = "END"
	// response keywords
	kwOK = "OK"
	// statement keywords
	kwUSERNAME = "USERNAME"
	kwPASSWORD = "PASSWORD"
	kwLOGOUT   = "LOGOUT"
	kwLIST     = "LIST"
	kwGET      = "GET"
	kwINSTCMD  = "INSTCMD"
	// branch keywords
	kwUPS = "UPS"
	kwVAR = "VAR"
	// UPS status flags
	FlagOnline = "OL"
	FlagOnBat  = "OB"
	FlagLowBat = "LB"
	FlagBadBat = "RB"
	FlagCharg  = "CHRG"
	FlagBypass = "BYPASS"
	// variable names
	varVOLTAGE    = "input.voltage"
	varTRANSFERHI = "input.transfer.high"
	varTRANSFERLO = "input.transfer.low"
	varUPSSTATUS  = "ups.status"
	// command names
	cmdLOADOFFDELAY = "load.off.delay"
	// Regexes
	reLISTLINE3F = `(\S+)\s(\S+)\s\"([^\"]+)\"`
	reLISTLINE4F = `(\S+)\s(\S+)\s(\S+)\s\"([^\"]+)\"`
	// error descriptions
	errLOGINFAIL = "login failure: %s"
	// notify texts
	NOTIFYTEXT = "Power loss: initiating shutdown!"
)
