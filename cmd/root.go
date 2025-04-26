/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "wotop",
	Short: "WOTOP is an open‑source Go framework designed to accelerate backend development with modern architectural patterns.",
	Long: `
			It brings together:

			- Clean Architecture for strict separation of business logic and infrastructure layers
			- Domain‑Driven Design (DDD) to model complex domains with clarity
			- Event‑Driven Microservices for loosely‑coupled, asynchronous communication
			- Cloud‑Native Microservices optimized for containerized, orchestrated environments

			Additionally, WOTOP integrates core patterns and tools out of the box:
			
			- CQRS (Command Query Responsibility Segregation)
			- RabbitMQ message broker
			- Event Sourcing for append‑only event storage
		`,

	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.wotop.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
