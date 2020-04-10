/*
Copyright © 2020 Leigh MacDonald <leigh.macdonald@gmail.com>

*/
package cmd

import (
	"bind_generator/app"
	"fmt"
	"github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"path"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "bind_generator",
	Short: "",
	Long:  ``,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		if viper.GetBool("debug") {
			log.SetLevel(log.DebugLevel)
		}
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
			bindsPath = "./binds.txt"
		}
		log.Debugf("console.log path: %s", consoleLogPath)
		log.Debugf("cfg path: %s", cfgPath)
		log.Debugf("binds path: %s", bindsPath)
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
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./config.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("debug", "d", false, "Enable debug output")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".bind_generator" (without extension).
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		log.Debugln("Using config file:", viper.ConfigFileUsed())
	}
}
