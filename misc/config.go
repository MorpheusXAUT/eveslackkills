package misc

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

// Configuration stores all configuration values required by the application
type Configuration struct {
	// DatabaseType represents the database type to be used as a backend
	DatabaseType int
	// DatabaseHost represents the hostname:port of the database backend
	DatabaseHost string
	// DatabaseSchema represents the schema/collection of the database backend
	DatabaseSchema string
	// DatabaseUser represents the username used to authenticate with the database backend
	DatabaseUser string
	// DatabasePassword represents the password used to authenticate with the database backend
	DatabasePassword string
	// DebugLevel represents the debug level for log messages
	DebugLevel int
	// EVECorporationIDs represents the IDs of the EVE corporations to check for kill- or loss-mails
	EVECorporationIDs []int
	// IgnoredSystemIDs represents the IDs of the solar systems to ignore while looking for kill- or loss-mails
	IgnoredSystemIDs []int
	// SlackWebhookURL represents the webhook URL provided by slack, used by the application to send chat messages
	SlackWebhookURL string
}

// LoadConfig creates a Configuration by either using commandline flags or a configuration file, returning an error if the parsing failed
func LoadConfig() (*Configuration, error) {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: eveslackkills [options]\n")
		flag.PrintDefaults()
		os.Exit(2)
	}

	flag.Parse()

	config, err := ParseJSONConfig(*configFileFlag)
	config = ParseCommandlineFlags(config)

	return config, err
}

// ParseJSONConfig parses a Configuration from a JSON encoded file, returning an error if the process failed
func ParseJSONConfig(path string) (*Configuration, error) {
	configFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	var config *Configuration

	err = json.NewDecoder(configFile).Decode(&config)

	return config, err
}
