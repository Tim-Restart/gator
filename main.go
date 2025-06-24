package main

import (
	"fmt"
	"gator/internal/config"
	"log"
	"os"
	"github.com/google/uuid"
)

type state struct {
	config		*config.Config
}

// Holds the command and arguments
type command struct {
	cmd			string
	args		[]string
}

type commands struct {
	cmds 		map[string]func(*state, command) error
}

func (c *commands) run(s *state, cmd command) error {
	handler, ok := c.cmds[cmd.cmd] // Something here about running...?
	if !ok {
		return fmt.Errorf("Unknown command: %s", cmd.cmd)
	}
	return handler(s, cmd)
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.cmds[name] = f
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) == 0  {
		return fmt.Errorf("Login details empty")
	}
	err := s.config.SetUser(cmd.args[0])
	if err != nil {
		return fmt.Errorf("Error setting user")
		os.Exit(1)
	}
	fmt.Printf("Current user %v\n", cmd.args[0])
	return nil
}




func main() {

	

	cfg, err := config.Read()
	if err != nil {
		log.Print("Error reading file")
		return
	}

	appState := state{
		config: &cfg,
	}

	

	cmds := commands{
		cmds: make(map[string]func(*state, command) error),
	}
	
	cmds.register("login", handlerLogin)

	// Using command line arguements os.Args 

	if len(os.Args) < 2 {
		fmt.Println("Command or username not provided")
		os.Exit(1)
	}
	userCommand := os.Args[1]
	fmt.Printf("User Command: %v\n", userCommand)
	var commandArgs []string
	if len(os.Args) > 2 {
		commandArgs = os.Args[2:]
	}

	cmd := command{
		cmd: userCommand,
		args: commandArgs,
	}
		
	err = cmds.run(&appState, cmd)
	if err != nil {
		fmt.Println("Error executing login")
		os.Exit(1)
	}


	cfg, err = config.Read()
	fmt.Println(cfg)
	return

}