package main

import (
	"fmt"
	"gator/internal/config"
	"log"
)

type state struct {
	config			*Config
}

// Holds the command and arguments
type command struct {
	cmd			string
	args		[]string
}

type commands struct {
	cmds 		map[string]func(*state, command) error
}

func (c *sommands) run(s *state, cmd command) error {
	handler, ok := c.cmds[cmd.cmd] // Something here about running...?
	if !ok {
		return fmt.Errorf("Unknown command: %s", cmd.cmd)
	}
	return handler(s, cmd)
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.cmds[name] = f
}


func main() {

	cfg, err := config.Read()
	if err != nil {
		log.Print("Error reading file")
		return
	}

	state := state{
		config: cfg,
	}

	comds := config.commands
	

	

	/*
	err = cfg.SetUser("Tim")
	if err != nil {
		log.Print("Error setting username")
		return
	}
		*/

	

	cfg, err = config.Read()
	fmt.Println(cfg)
	return

}