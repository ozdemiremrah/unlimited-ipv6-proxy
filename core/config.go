package core

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
)

var (
	// ConfigFileDir the directory should be in located.
	ConfigFileDir = "src"
)

// DatabaseConfig model
type ProxyConfig struct {
	Host      string `json:"host"`
	Port      uint16 `json:"port"`
	TempIP    string `json:"temp_ip"`
	Interface string `json:"interface"`
	SubnetMask uint16 `json:"subnet"`
}

// Config model
type Config struct {
	Version string      `json:"version"`
	Proxy   ProxyConfig `json:"proxy"`
}

var version = "1.0.0"

var defaultConfig = `{
	"version": "%s",
	"proxy": {
		"host": "127.0.0.1",
		"port": 8080,
		"temp_ip": "::",
		"interface" : "ens3",
		"subnet" : 64
	}
}
`

var config Config

func getConfigFilePath() (string, error) {
	rootDir, err := os.Getwd()

	if err != nil {
		return "", err
	}

	filepath := path.Join(rootDir, "config.json")

	return filepath, nil
}

func createDefaultConfigFile(path string) error {
	file, err := os.Create(path)

	defer file.Close()

	if err != nil {
		return err
	}

	content := fmt.Sprintf(defaultConfig, version)
	bytes := []byte(content)

	if err := ioutil.WriteFile(path, bytes, 0644); err != nil {
		return err
	}

	return nil
}

// GetConfig returns the config object.
func GetConfig() *Config {
	return &config
}

// ReadConfig reads the config file, and creates default config file if doesn't exist.
func ReadConfig() error {
	var path string
	var err error

	if path, err = getConfigFilePath(); err != nil {
		return err
	}

	fmt.Println(path)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err = createDefaultConfigFile(path); err != nil {
			return err
		}
	}

	bytes, err := ioutil.ReadFile(path)

	if err != nil {
		return err
	}

	if err = json.Unmarshal(bytes, &config); err != nil {
		return err
	}

	return nil
}
