package config

import (
	"os"
	"log"
)

const configFileName = ".gatorconfig.json"

type Config struct {
	DbURL           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

func getConfigFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Print("Error finding home directory")
		return "", err
	}
	return home, nil
} 

func Read() Config {
	
	home := getConfigFilePath()
	configFile := home + configFileName

	reader, err := os.Readfile(configFile)
	if err != nil {
		log.Print("Error reading file at home directory")
		return err
	}
}

func (c *Config) SetUser(user string) {
	c.CurrentUserName = user
}