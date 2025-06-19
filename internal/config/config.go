package config

import (
	"os"
	"log"
	"filepath"
	"encoding/json"
	"fmt"
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
	configFile := filepath.Join(home, configFileName)
	return configFile, nil
} 

func Read() (Config, error) {
	c := Config{}

	configFile, err := getConfigFilePath()
	if err != nil {
		log.Print("Error getting file path")
		return Config{}, err
	}
	reader, err := os.Readfile(configFile)
	if err != nil {
		log.Print("Error reading file at home directory")
		return Config{}, err
	}
	
	err = json.Unmarshal(reader, &c)
	if err != nil {
		log.Print("Error unmarshalling JSON file")
		return Config{}, err
	}

	return c, nil

}

func (c *Config) SetUser(user string) error {
	c.CurrentUserName = user
	if c.CurrentUserName == "" {
		return fmt.Errorf("Error setting username")
	}
	return nil
}