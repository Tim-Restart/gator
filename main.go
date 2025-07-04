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
	cmds.register("feeds", handlerFeeds)
	// Handlers that require login
	cmds.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	cmds.register("follow", middlewareLoggedIn(handlerFollow))
	cmds.register("following", middlewareLoggedIn(handlerFollowing))
	cmds.register("unfollow", middlewareLoggedIn(handlerUnfollow))

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

// Helper functions

func getUserId(s *state, user string) (uuid.UUID, error) {
	
	var zeroID uuid.UUID
	ctx := context.Background()
	userName, err := s.db.GetUser(ctx, user)
	if errors.Is(err, sql.ErrNoRows) {
		return zeroID, err
	} else if err != nil {
		return zeroID, err
	}
	return userName.ID, nil
}

/*
Fix this later

func newFeedFollow(s *state, user_id uuid.UUID, feed_id uuid.UUID) error {
newFollow := database.CreateFeedFollowParams{
		ID: uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID: id, 
		FeedID: feed.ID,
	}

	feedFollow, err := s.db.CreateFeedFollow(ctx, newFollow)
	if err != nil {
		fmt.Println("Error creating new feed follow")
		os.Exit(1)
	}
	return nil
}

*/

// Middleware function

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	// Handle the logged in check here
	// Takes a handler
	return func(s *state, cmd command) error {
		ctx := context.Background()
		user := s.config.CurrentUserName

		userName, err := s.db.GetUser(ctx, user)
		if errors.Is(err, sql.ErrNoRows) {
			fmt.Println("User doesn't exist")
			return err
		} else if err != nil {
			fmt.Println("Error checking existing user")
			return err
		} 
		return handler(s, cmd, userName)

	}
}



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
	fmt.Printf("New User added, Welcome %v\n", newUser)
	fmt.Printf("ID: %v\n", newArgs.ID)
	fmt.Printf("Created At: %v\n", newArgs.CreatedAt)
	fmt.Printf("Updated At: %v\n", newArgs.UpdatedAt)
	fmt.Printf("Name: %v\n", newArgs.Name)


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

func handlerAddFeed(s *state, cmd command, user database.User) error {
	ctx := context.Background()

	var name string
	var url string
	

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
		UserID: user.ID,
	}

	feed, err := s.db.CreateFeed(ctx, newFeed)
	if err != nil {
		fmt.Println("Error creating feed")
		os.Exit(1)
	}

	fmt.Println(feed)

	// Add a feed follow here - refactor later this to a helper function

	newFollow := database.CreateFeedFollowParams{
		ID: uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID: user.ID, 
		FeedID: feed.ID,
	}

	feedFollow, err := s.db.CreateFeedFollow(ctx, newFollow)
	if err != nil {
		fmt.Println("Error creating new feed follow")
		os.Exit(1)
	}

	fmt.Printf("%v now following %v\n", user.Name, feedFollow.ID)


	return nil
}

func handlerFeeds(s *state, cmd command) error {
	ctx := context.Background()
	//var feed database.CreateFeedParams
	// Returns all the feeds
	feed, err := s.db.GetFeeds(ctx)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			// No users found return
			fmt.Println("No feeds found")
			return err
	}
		fmt.Println("Error getting feeds")
		return err	
	}

	// Name of the feed

	for _, feedItem := range feed {
		
		fmt.Printf("Name: %v\n", feedItem.Name)
		fmt.Printf("URL: %v\n", feedItem.Url)
		userId, err := s.db.GetUserName(ctx, feedItem.UserID)
		if err != nil {
			fmt.Println("Error retreiving user name")
			os.Exit(1)
		}
		fmt.Printf("User: %v\n", userId)
		}
	
	return nil

}


func handlerFollow(s *state, cmd command, user database.User) error {
	ctx := context.Background()
	var urlID uuid.UUID
	var url string


	// Checks to make sure there is a cmd argument and assigns it to URL
	if len(cmd.args) > 0 {
		url = cmd.args[0]
	} else {
		fmt.Println("No url provided")
		os.Exit(1)
	}

	feedURL, err := s.db.GetFeedUrl(ctx, url)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			// No users found create a new record
			fmt.Println("URL not found, creating new record")

			err = handlerAddFeed(s, cmd, user)
			if err != nil {
				fmt.Println("Error creating new feed")
				return err
			}
			feedURL, err := s.db.GetFeedUrl(ctx, url)
			if err != nil {
				fmt.Println("At least we aren't going round in circles...")
			}
			urlID = feedURL.ID

			} else {
				fmt.Println("Error getting URL")
				return err
			}
	}

	urlID = feedURL.ID

	newFollow := database.CreateFeedFollowParams{
		ID: uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID: user.ID, 
		FeedID: urlID,
	}

	feedFollow, err := s.db.CreateFeedFollow(ctx, newFollow)
	if err != nil {
		fmt.Println("Error creating new feed follow")
		os.Exit(1)
	}

	fmt.Printf("Feed Name: %v\n", feedFollow.FeedName)
	fmt.Printf("Feed User: %v\n", user.Name)
	return nil
}

func handlerFollowing(s *state, cmd command, user database.User) error {
	ctx := context.Background()

	follows, err := s.db.GetFeedFollowsForUser(ctx, user.ID)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			// No users found create a new record
			fmt.Println("no feed follows for this user")
			os.Exit(1)
		} else {
			fmt.Println("Error getting feed follows for user")
			os.Exit(1)
		}
	}

	for i, feed := range follows {
		fmt.Printf("%v. Feeds being followed: %v", (i + 1), feed.FeedName)
	}
	return nil

}	

func handlerUnfollow(s *state, cmd command, user database.User) error {
	ctx := context.Background()
	var url string

	if len(cmd.args) > 0 {
		url = cmd.args[0]
	} else {
		fmt.Println("No url provided")
		os.Exit(1)
	}

	feed, err := s.db.GetFeedUrl(ctx, url)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			// No users found create a new record
			fmt.Println("URL not found, cannot unfollow")
		} else {
			fmt.Println("Error getting Feed Details")
			return err
		}
	}

	feedUnfollow := database.DeleteFeedFollowsParams{
		UserID: user.ID,
		FeedID: feed.ID,
	}
	err = s.db.DeleteFeedFollows(ctx, feedUnfollow)
	if err != nil {
		fmt.Println("Error unfollowing feed")
		return err
	}
	fmt.Printf("%v no longer following: %v", user.Name, feed.Name)
	return nil
}

func scrapeFeeds(s *state) error {
	ctx := context.Background()
	
	feedID, err := s.db.GetNextFeedToFetch(ctx)
	if err.Error() == "sql: no rows in result set" {
			// No users found create a new record
			fmt.Println("No feeds found being followed, grow your user base")
	} else {
		fmt.Println("Error getting next Feed Details")
		return err
	}

	feedInfo, err := s.db.GetFeedURLfromID(ctx, feedID)
	if err.Error() == "sql: no rows in result set" {
			// No users found create a new record
			fmt.Println("No feeds found with given ID")
	} else {
		fmt.Println("Error getting feed information")
		return err
	}

	markReturns := database.MarkFeedFetchedParams{
		UpdatedAt: time.Now(),
		LastFetchedAt: sql.NullTime{
			Time: time.Now(),
			Valid: true,
		},
		ID: feedID,
	}

	err = s.db.MarkFeedFetched(ctx, markReturns)
	if err != nil {
		fmt.Println("Error marking feed as fetched")
		return err
	}

	feed, err := s.db.GetFeedUrl(ctx, feedInfo.Url)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			// No users found create a new record
			fmt.Println("URL not found")
			return err
		} else {
			fmt.Println("Error getting Feed Details")
			return err
		}
	}

	response, err := fetchFeed(ctx, feed.Url)
	if err != nil {
		fmt.Println("Error fetching feed")
		return err
	}

	fmt.Println(response)
	return nil

}
