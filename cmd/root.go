/*
Copyright © 2020 Leigh MacDonald <leigh.macdonald@gmail.com>

*/
package cmd

import (
	"bind_generator/app"
	"bind_generator/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"path"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "bind_generator",
	Short: "",
	Long:  ``,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		consoleLogPath := viper.GetString("consoleLogPath")
		if consoleLogPath == "" {
			consoleLogPath = path.Join(os.Getenv("ProgramFiles(x86)"), "Steam", "steamapps", "common", "Team Fortress 2", "tf", "console.log")
		}
		cfgPath := viper.GetString("cfgPath")
		if cfgPath == "" {
			consoleLogPath = path.Join(os.Getenv("ProgramFiles(x86)"), "Steam", "steamapps", "common", "Team Fortress 2", "tf", "cfg", "bind_generator.cfg")
		}
		bindsPath := viper.GetString("bindsPath")
		if bindsPath == "" {
			bindsPath = "./binds_custom.txt"
		}
		app := app.New(consoleLogPath, bindsPath, cfgPath)
		app.Start()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Errorf("Error executing command: %s", err.Error())
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(config.Load)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&config.CfgFile, "config", "", "config file (default is ./config.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("debug", "d", false, "Enable debug output")
}
