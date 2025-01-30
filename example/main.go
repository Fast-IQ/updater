package main

import (
	"fmt"
	"github.com/Fast-IQ/updater"
)

func main() {
	url := "127.0.0.1:8080"
	update := updater.NewUpdater()
	err := update.UpdateFile(url, "test.txt")
	if err != nil {
		fmt.Println("err:", err)
	}

}
