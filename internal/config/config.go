package config

import (
	"os"
	"log"
	"path/filepath"
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
	cfg := Config{}

	configFile, err := getConfigFilePath()
	if err != nil {
		log.Print("Error getting file path")
		return Config{}, err
	}
	reader, err := os.ReadFile(configFile)
	if err != nil {
		log.Print("Error reading file at home directory")
		return Config{}, err
	}
	
	err = json.Unmarshal(reader, &cfg)
	if err != nil {
		log.Print("Error unmarshalling JSON file")
		return Config{}, err
	}

	return cfg, nil

}

func (cfg *Config) SetUser(user string) error {
	cfg.CurrentUserName = user
	if cfg.CurrentUserName == "" {
		return fmt.Errorf("Error setting username")
	}
	err := write(cfg)
	if err != nil {
		log.Print("Error setting username")
		return err
	}
	return nil
}

func write(cfg *Config) error {
	
	body, err := json.Marshal(*cfg) // I'm not sure if thats what I actually want to marshall??
	if err != nil {
		log.Print("Error marshalling JSON")
		return err
	}
	// Add the body into something here, like a writer?
	configFile, err := getConfigFilePath()
	if err != nil {
		log.Print("Error getting file path")
		return err
	}
	

	err = os.WriteFile(configFile, body, 0666)
	if err != nil {
		log.Print("Error writing file to path")
		return err
	}
	
	return nil
}