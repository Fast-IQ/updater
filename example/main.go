package main

import (
	"fmt"
	"github.com/Fast-IQ/updater-client"
)

func main() {
	update := updater_client.NewUpdater()
	err := update.UpdateFile("http://vwDevLock.sima-land.local:32131", "example/test.txt")
	if err != nil {
		fmt.Println("err:", err)
	}

}
