package main

import (
	"fmt"
	"github.com/Fast-IQ/updater-client"
)

func main() {
	update := updater_client.NewUpdater()
	err := update.UpdateFile(url, "example/test.txt")
	if err != nil {
		fmt.Println("err:", err)
	}

}
