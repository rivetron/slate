package cmd

import (
	"fmt"
	"os"

	"github.com/rivetron/slate/src/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
	Long:  "Manage slate configuration files.",
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Create a default configuration file",
	Long: `Create a default configuration file at ~/.config/slate/slate.yaml
	
This will create a configuration file with sensible defaults that you can customize.`,
	Run: func(cmd *cobra.Command, args []string) {
		path, err := config.CreateDefaultConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
			os.Exit(1)
		}

		fmt.Printf("Created default config at: %s\n", path)
		fmt.Println("You can now edit this file to customize your presentation settings.")
	},
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Create a default configuration file",
	Long:  "Create a default configuration file at ~/.config/slate/slate.yaml",
	Run: func(cmd *cobra.Command, args []string) {
		loader := config.New()
		cfg, err := loader.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
			os.Exit(1)
		}

		fmt.Println("Current Configuration:")
		fmt.Println("=====================")
		fmt.Printf("\nTheme:\n")
		fmt.Printf("  Mode: %s\n", cfg.Theme.Mode)
		fmt.Printf("  Glamour Style: %s\n", cfg.Theme.GlamourStyle)
		fmt.Printf("  Show Progress: %v\n", cfg.Theme.ShowProgress)
		fmt.Printf("  Show Slide Number: %v\n", cfg.Theme.ShowSlideNum)

		fmt.Printf("\nPresentation:\n")
		fmt.Printf("  Word Wrap: %d\n", cfg.Presentation.WordWrap)
		fmt.Printf("  Margin: %d\n", cfg.Presentation.Margin)
		fmt.Printf("  Padding: %d\n", cfg.Presentation.Padding)

		fmt.Printf("\nKeybindings:\n")
		fmt.Printf("  Next: %v\n", cfg.Keybindings.Next)
		fmt.Printf("  Previous: %v\n", cfg.Keybindings.Previous)
		fmt.Printf("  First: %v\n", cfg.Keybindings.First)
		fmt.Printf("  Last: %v\n", cfg.Keybindings.Last)
		fmt.Printf("  Quit: %v\n", cfg.Keybindings.Quit)

		if configPath := loader.GetConfigPath(); configPath != "" {
			fmt.Printf("\nConfig file: %s\n", configPath)
		} else {
			fmt.Printf("\nUsing default configuration (no config file found)\n")
		}
	},
}

var configPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Show configuration file path",
	Long:  "Show the path to the configuration file being used",
	Run: func(cmd *cobra.Command, args []string) {
		loader := config.New()
		path, err := loader.FindConfig()
		if err != nil {
			fmt.Println("No configuration file found.")
			fmt.Println("Run 'slate config init' to create one.")
			return
		}

		fmt.Println(path)
	},
}

var configExampleCmd = &cobra.Command{
	Use:   "example",
	Short: "Show an example configuration",
	Long:  "Display an example configuration file with all available options.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Example Configuration:")
		fmt.Println("=====================")
		fmt.Println()
		fmt.Print(config.ExampleConfig())
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configPathCmd)
	configCmd.AddCommand(configExampleCmd)
}
