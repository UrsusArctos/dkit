package main

import (
	"fmt"
	"net"
	"os"

	"github.com/UrsusArctos/dkit/pkg/upsdclient"
)

func main() {
	UC, err := upsdclient.NewConnection(net.TCPAddr{IP: net.ParseIP("10.0.0.1"), Port: 3493})
	if err == nil {
		pwd, _ := os.ReadFile("password.txt")
		UC.Login("admin", string(pwd))
		fmt.Println("Mains = ", UC.IsPowerMainsGood())
		fmt.Printf("[ %s ]= %t\n", upsdclient.FlagOnline, UC.GetStatusOn(upsdclient.FlagOnline))
		fmt.Printf("[ %s ]= %t\n", upsdclient.FlagLowBat, UC.GetStatusOn(upsdclient.FlagLowBat))
		fmt.Printf("[ %s ]= %t\n", upsdclient.FlagCharg, UC.GetStatusOn(upsdclient.FlagCharg))
		fmt.Printf("[ %s ]= %t\n", upsdclient.FlagOnBat, UC.GetStatusOn(upsdclient.FlagOnBat))
		UC.CloseConnection()

	} else {
		fmt.Printf("Unable to connect: %+v\n", err)
	}
}
