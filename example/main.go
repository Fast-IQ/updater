package main

import (
	"fmt"
	"github.com/Fast-IQ/updater"
)

func main() {
	update := updater.NewUpdater()
	err := update.UpdateFile(url, "example/test.txt")
	if err != nil {
		fmt.Println("err:", err)
	}

}
