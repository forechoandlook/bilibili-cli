package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/forechoandlook/bilibili-cli/internal/commands"
)

var version = "0.6.2"

func main() {
	var verbose bool

	rootCmd := &cobra.Command{
		Use:     "bili",
		Short:   "bili — Bilibili CLI tool 📺",
		Version: version,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if verbose {
				// We could setup verbose logging here
			}
		},
	}

	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable debug logging.")

	// Add subcommands
	commands.AddAuthCommands(rootCmd)
	commands.AddVideoCommands(rootCmd)
	commands.AddUserCommands(rootCmd)
	commands.AddDiscoveryCommands(rootCmd)
	commands.AddCollectionCommands(rootCmd)
	commands.AddInteractionCommands(rootCmd)

	// Custom panic handler to catch simulated `sys.exit` equivalents
	defer func() {
		if r := recover(); r != nil {
			if str, ok := r.(string); ok && len(str) >= 4 && str[:4] == "exit" {
				if len(str) > 5 {
					os.Exit(1)
				}
				os.Exit(0)
			}
			panic(r)
		}
	}()

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
