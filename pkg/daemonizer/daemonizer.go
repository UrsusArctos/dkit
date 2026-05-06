//go:build linux

package daemonizer

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	/*
	   	The conventions for daemons are as such:

	   --conf /full/path/to/daemon.conf
	   or defaults /etc/name.conf
	   --pid /full/path/to/daemon.pid
	   or defaults to /run/name.pid
	   --logpath /full/path/to				(nb: no filename!)
	   or defaults to /var/log (assuming name.log.gz there)
	   --foreground
	   or defaults to false (implying being started as a daemon)
	*/
	// Config file
	optConf  = "conf"
	defConf  = "/etc/%s.conf"
	descConf = "Full pathname of config file"
	// PID file
	optPID  = "pid"
	defPID  = "/run/%s.pid"
	descPID = "Full pathname of pid file"
	// Log path
	optLog  = "logpath"
	defLog  = "/var/log/"
	descLog = "Directory for log files"
	// Foreground
	optFore  = "foreground"
	defFore  = false
	descFore = "Start in a foreground mode (don't daemonize)"
)

type (
	TLinuxDaemon struct {
		// internal
		name    string
		pidFile string
		// exported
		Foreground bool
		LogPath    string
		ConfFile   string
		// handlers
		FuncInit  TDaemonCycle
		FuncClose TDaemonCycle
		FuncMain  TDaemonCycle
	}

	TDaemonCycle func() (err error)
)

func NewLinuxDaemon(dname string) (ld TLinuxDaemon) {
	ld.name = dname
	ld.parseCmdLine()
	ld.writePidFile()
	ld.FuncInit = nil
	ld.FuncClose = nil
	ld.FuncMain = nil
	return ld
}

func (ld *TLinuxDaemon) Close() {
	os.Remove(ld.pidFile)
}

func (ld *TLinuxDaemon) parseCmdLine() {
	flag.StringVar(&ld.ConfFile, optConf, fmt.Sprintf(defConf, ld.name), descConf)
	flag.StringVar(&ld.pidFile, optPID, fmt.Sprintf(defPID, ld.name), descPID)
	flag.StringVar(&ld.LogPath, optLog, defLog, descLog)
	flag.BoolVar(&ld.Foreground, optFore, defFore, descFore)
	flag.Parse()
}

func (ld TLinuxDaemon) writePidFile() {
	f, err := os.Create(ld.pidFile)
	if err == nil {
		defer f.Close()
		f.WriteString(fmt.Sprintf("%d", os.Getpid()))
	}
}

func (ld TLinuxDaemon) Run(interval time.Duration) error {
	// 1. Prepare the event loop

	// 1.2. prepare the signal handling
	kill := make(chan os.Signal, 1)
	signal.Notify(kill, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(kill)

	// 1.3. prepare the ticker
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// 2. Three main phases; Skip each entirely if not defined.

	// 2.1. Initialization handler
	if ld.FuncInit != nil {
		if err := ld.FuncInit(); err != nil {
			return err
		}
	}

	// 2.2. Event loop
	var errMain error = nil
	if ld.FuncMain != nil {
	EventLoop:
		for errMain == nil {
			select {
			case <-ticker.C:
				errMain = ld.FuncMain()
			case <-kill:
				break EventLoop
			}
		}
	}

	// 2.3. Finalization handler
	if ld.FuncClose != nil {
		if err := ld.FuncClose(); err != nil {
			return err
		}
	}

	return errMain
}
