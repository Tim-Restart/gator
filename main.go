package main

import _ "github.com/lib/pq"

import (
	"fmt"
	"gator/internal/config"
	"log"
	"os"
	"github.com/google/uuid"
	"gator/internal/database"
	"time"
	"context"
	"database/sql"
	"errors"
)

type state struct {
	config		*config.Config
	db			*database.Queries
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






func main() {


	cfg, err := config.Read()
	if err != nil {
		log.Print("Error reading file")
		return
	}

	log.Printf("cfg.DbURL: %v", cfg.DbURL)

	db, err := sql.Open("postgres", cfg.DbURL)
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
    // ...continue with program, using db
	dbQueries := database.New(db)
	

	

	appState := state{
		config: &cfg,
		db: dbQueries,
	}



	

	cmds := commands{
		cmds: make(map[string]func(*state, command) error),
	}
	
	// Register command handlers here

	cmds.register("login", handlerLogin)
	cmds.register("register", handlerRegister)
	cmds.register("reset", handlerResetDB)
	cmds.register("users", handleGetUsers)
	cmds.register("agg", handlerAgg)
	cmds.register("addfeed", handlerAddFeed)

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
		fmt.Println("Error executing command")
		fmt.Print(err)
		os.Exit(1)
	}


	
	return

}

// Main function closed
// Handlers below

func handlerLogin(s *state, cmd command) error {
	ctx := context.Background()
	if len(cmd.args) == 0  {
		return fmt.Errorf("Login details empty")
	}
	newUser := cmd.args[0]
	userReturned, err := s.db.GetUser(ctx, newUser)
	if err != nil {
		log.Printf("This is the error: %v", err)
		fmt.Print("Error checking existing user\n")
		return err
	}
	user := userReturned.Name
	if user == "" {
		fmt.Print("User doesn't exist")
		os.Exit(1)
	}

	err = s.config.SetUser(cmd.args[0])
	if err != nil {
		return fmt.Errorf("Error setting user")
		os.Exit(1)
	}
	fmt.Printf("Current user %v\n", cmd.args[0])
	return nil
}

func handlerRegister(s *state, cmd command) error {
	ctx := context.Background()
	if len(cmd.args) == 0  {
		return fmt.Errorf("register details empty")
	}
	
	newUser := cmd.args[0]
	_, err := s.db.GetUser(ctx, newUser)
	if errors.Is(err, sql.ErrNoRows) {
		log.Print("Welcome new user")
	} else if err != nil {
		fmt.Print("Error checking existing user")
		return err
	} else {
		fmt.Println("Username already exists")
		os.Exit(1)
	}
		

	newArgs := database.CreateUserParams{
		ID: uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name: newUser,
	}

	_, err = s.db.CreateUser(ctx, newArgs)
	if err != nil {
		fmt.Println("Unable to create new user")
		os.Exit(1)
	}
	err = s.config.SetUser(newUser)
	if err != nil {
		return fmt.Errorf("Error setting user")
		os.Exit(1)
	}
	log.Printf("New User added, Welcome %v\n", newUser)
	log.Printf("ID: %v\n", newArgs.ID)
	log.Printf("Created At: %v\n", newArgs.CreatedAt)
	log.Printf("Updated At: %v\n", newArgs.UpdatedAt)
	log.Printf("Name: %v\n", newArgs.Name)


	return nil
	
}

func handlerResetDB(s *state, cmd command) error {
	ctx := context.Background()
	err := s.db.DeleteUsers(ctx)
	if err != nil {
		return err
	}
	fmt.Println("Database reset complete")
	return nil
}

func handleGetUsers(s *state, cmd command) error {
	ctx := context.Background()

		
	// GetUsers command for all users
	// Will return a struct full of the users
	// Need to print to console

	users, err := s.db.GetAllUsers(ctx)
	if err != nil {

		if err.Error() == "sql: no rows in result set" {
			// No users found return
			fmt.Println("No users found")
			return err
		}
	}

	for _, user := range users {
		
		if user.Name == s.config.CurrentUserName {
		fmt.Printf("* %v (current)\n", user.Name)
		} else {
			fmt.Printf("* %v\n", user.Name)
		}
		
	}
	return nil
}

func handlerAgg(s *state, cmd command) error {
	ctx := context.Background()
	var url string

	if len(cmd.args) > 0 {
		url = cmd.args[0]
	} else {
		url = "https://www.wagslane.dev/index.xml"
	}

	response, err := fetchFeed(ctx, url)
	if err != nil {
		fmt.Println("Error fetching feed")
		fmt.Printf("Error: %v", err)
		return err
	}

	fmt.Println(response)
	return nil
}

func handlerAddFeed(s *state, cmd command) error {
	ctx := context.Background()

	var name string
	var url string
	var id uuid.UUID

	user := s.config.CurrentUserName

	userName, err := s.db.GetUser(ctx, user)
	if errors.Is(err, sql.ErrNoRows) {
		log.Print("User doesn't exist")
	} else if err != nil {
		fmt.Print("Error checking existing user")
		return err
	} else {
		id = userName.ID
	}

	if len(cmd.args) > 0 {
		name = cmd.args[0]
	} else {
		fmt.Println("No name provided")
		os.Exit(1)
	}

	if len(cmd.args) > 1 {
		url = cmd.args[1]
	} else {
		fmt.Println("No url provided")
		os.Exit(1)
	}

	newFeed := database.CreateFeedParams{
		ID: uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name: name,
		Url: url,
		UserID: id,
	}

	feed, err := s.db.CreateFeed(ctx, newFeed)
	if err != nil {
		fmt.Println("Error creating feed")
		os.Exit(1)
	}

	fmt.Println(feed)
	return nil
}


