package main

import (
	"fmt"
	"gator/internal/config"
	"log"
)

func main() {

	cfg, err := config.Read()
	if err != nil {
		log.Print("Error reading file")
		return
	}
	err = cfg.SetUser("Tim")
	if err != nil {
		log.Print("Error setting username")
		return
	}
	cfg, err = config.Read()
	fmt.Println(cfg)
	return

}