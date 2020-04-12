package config

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var CfgFile string

// initConfig reads in config file and ENV variables if set.
func InitConfig() {
	if CfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(CfgFile)
	} else {
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv() // read in environment variables that match
	log.SetFormatter(&log.TextFormatter{
		ForceColors: true,
	})
	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		if viper.GetBool("debug") {
			log.SetLevel(log.DebugLevel)
		}
		log.Infof("Using config file: %s", viper.ConfigFileUsed())
	}
}
