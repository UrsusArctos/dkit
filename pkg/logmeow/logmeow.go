//go:build linux

package logmeow

import (
	"compress/gzip"
	"fmt"
	"log/syslog"
	"os"
	"time"
)

const (
	// Facility flags
	FacConsole = 1 << 0
	FacFile    = 1 << 1
	FacSyslog  = 1 << 2
	// Severity mapping
	logSevInfo    = syslog.LOG_INFO
	logSevWarning = syslog.LOG_WARNING
	logSevError   = syslog.LOG_ERR
	// Color indices
	colorRed    = 1
	colorYellow = 3
	colorBlue   = 4
	// colorGreen   = 2
	// colorMagenta = 5
	// colorCyan    = 6
	// colorWhite   = 7
)

type (
	TLogMeow struct {
		enabledFacilities uint8
		name              string
		logfile           *os.File
		gzwr              *gzip.Writer
		syslog            *syslog.Writer
	}
)

func timePrefix() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func severityColor(sev syslog.Priority) uint8 {
	switch sev {
	case logSevInfo:
		return 0
	case logSevWarning:
		return colorYellow
	case logSevError:
		return colorRed
	default:
		return 0
	}
}

func enColor(text string, colorindex uint8) string {
	if colorindex != 0 {
		return fmt.Sprintf("\033[1;%dm%s\033[0m", (30 + colorindex), text)
	} else {
		return text
	}
}

func NewLogMeow(meowname string, enfac uint8, auxargs ...string) (lm TLogMeow) {
	lm.name = meowname
	// console
	if (enfac & FacConsole) != 0 {
		// Console does not need special initialization
		lm.enabledFacilities |= FacConsole
	}
	// log file
	if (enfac & FacFile) != 0 {
		defpath := "./"
		if len(auxargs) > 0 {
			defpath = auxargs[0]
		}
		lm.logfile, _ = os.OpenFile(fmt.Sprintf("%s%s.log.gz", defpath, lm.name), os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0640)
		lm.gzwr = gzip.NewWriter(lm.logfile)
		lm.enabledFacilities |= FacFile
	}
	// syslog
	if (enfac & FacSyslog) != 0 {
		lm.syslog, _ = syslog.Dial("", "", syslog.LOG_INFO|syslog.LOG_DAEMON, lm.name)
		lm.enabledFacilities |= FacSyslog
	}
	return lm
}

func (meow TLogMeow) IsFacilityEnabled(fac uint8) bool {
	return (meow.enabledFacilities & fac) != 0
}

func (meow *TLogMeow) Close() {
	// again, console does not need cleanup
	// log file
	if meow.IsFacilityEnabled(FacFile) {
		meow.gzwr.Flush()
		meow.gzwr.Close()
		meow.logfile.Close()
	}
	// syslog
	if meow.IsFacilityEnabled(FacSyslog) {
		meow.syslog.Close()
	}
}

func (meow TLogMeow) LogEventInfo(edesc string) {
	meow.logEventCommon(edesc, logSevInfo)
}

func (meow TLogMeow) LogEventWarning(edesc string) {
	meow.logEventCommon(edesc, logSevWarning)
}

func (meow TLogMeow) LogEventError(edesc string) {
	meow.logEventCommon(edesc, logSevError)
}

func (meow TLogMeow) logEventCommon(eventdescr string, severity syslog.Priority) {
	// console (timeprefix : YES, coloring : YES, addLF : YES)
	if meow.IsFacilityEnabled(FacConsole) {
		fmt.Printf("%s %s\n", enColor(timePrefix(), colorBlue), enColor(eventdescr, severityColor(severity)))
	}
	// log file (timeprefix : YES, coloring : NO, addLF : YES)
	if meow.IsFacilityEnabled(FacFile) {
		meow.gzwr.Write([]byte(fmt.Sprintf("%s %s\n", timePrefix(), eventdescr)))
	}
	// syslog (timeprefix : NO, coloring : NO, addLF : NO)
	if meow.IsFacilityEnabled(FacSyslog) {
		switch severity {
		case logSevInfo:
			meow.syslog.Info(eventdescr)
		case logSevWarning:
			meow.syslog.Warning(eventdescr)
		case logSevError:
			meow.syslog.Err(eventdescr)
		default:
			meow.syslog.Info(eventdescr)
		}
	}
}
