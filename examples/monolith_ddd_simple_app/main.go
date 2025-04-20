package main

import (
	"flag"
	"fmt"
	"github.com/a-aslani/wotop"
	"github.com/a-aslani/wotop/examples/monolith_ddd_simple_app/cmd"
	"github.com/a-aslani/wotop/examples/monolith_ddd_simple_app/configs"
	"os"
)

// Version specifies the current version of the application.
var Version = "0.0.1"

// main is the entry point of the application.
// It loads the configuration, initializes the application map, and runs the selected application.
func main() {

	// Retrieve the configuration file path from the environment variable CONFIG_FILE.
	configFile := os.Getenv("CONFIG_FILE")

	// If CONFIG_FILE is not set, default to ".env".
	if configFile == "" {
		configFile = ".env"
	}

	// Load the configuration from the specified file.
	cfg, err := configs.LoadConfig(configFile)
	if err != nil {
		// Print an error message if the configuration file cannot be loaded.
		fmt.Printf("config file error: %s", err.Error())
		return
	}

	// Define a map of application names to their corresponding runners.
	appMap := map[string]wotop.Runner[configs.Config]{
		"product": cmd.NewProduct(),
	}

	// Parse command-line flags.
	flag.Parse()

	// Retrieve the application name from the command-line arguments.
	app, exist := appMap[flag.Arg(0)]
	if !exist {
		// Print usage instructions if the application name is not found in the map.
		fmt.Printf("You may try :\n\n")
		for appName := range appMap {
			fmt.Printf("    go run main.go %s\n", appName)
		}
		fmt.Printf("\n")
		return
	}

	// Print the configuration file path and application version.
	fmt.Printf("Config: %s - Version: %s\n", configFile, Version)

	// Run the selected application with the loaded configuration.
	err = app.Run(cfg)
	if err != nil {
		// Print an error message if the application fails to run.
		fmt.Printf("run error: %s", err.Error())
		return
	}
}
