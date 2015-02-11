package misc

import "flag"

var (
	debugLevelFlag = flag.Int("debug", 3, "Sets the debug level (0-9), lower number displays more messages")
	configFileFlag = flag.String("config", "config.cfg", "Path to the config file to parse")
)

// ParseCommandlineFlags parses the command line flags used with the application
func ParseCommandlineFlags(config *Configuration) *Configuration {
	if *debugLevelFlag != 3 {
		config.DebugLevel = *debugLevelFlag
	}

	return config
}
