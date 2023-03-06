package main

import (
	"fmt"
	"net"
	"os"

	"github.com/UrsusArctos/dkit/pkg/upsdclient"
)

func main() {
	UC, cerr := upsdclient.NewConnection(net.TCPAddr{IP: net.ParseIP("10.0.0.1"), Port: 3493})
	if cerr == nil {
		pwd, _ := os.ReadFile(".creds/password-upsd.txt")
		lerr := UC.Login("admin", string(pwd))
		if lerr == nil {
			fmt.Println("Mains = ", UC.IsPowerMainsGood())
			fmt.Printf("[ %s ]= %t\n", upsdclient.FlagOnline, UC.GetStatusOn(upsdclient.FlagOnline))
			fmt.Printf("[ %s ]= %t\n", upsdclient.FlagLowBat, UC.GetStatusOn(upsdclient.FlagLowBat))
			fmt.Printf("[ %s ]= %t\n", upsdclient.FlagCharg, UC.GetStatusOn(upsdclient.FlagCharg))
			fmt.Printf("[ %s ]= %t\n", upsdclient.FlagOnBat, UC.GetStatusOn(upsdclient.FlagOnBat))
			UC.CloseConnection()
		} else {
			fmt.Printf("Unable to login: %+v\n", lerr)
		}
	} else {
		fmt.Printf("Unable to connect: %+v\n", cerr)
	}
}
