package config

import (
	"fmt"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
)

var (
	defaultPath         = "./config"
	defaultFileBaseName = "settings"
	defaultFileType     = "yaml"
)

// ViperInit viper init with run mode and envPrefix
func ViperInit(runMode string, envPrefix string) error {
	viper.SetConfigFile(getFilename("common"))
	if err := viper.ReadInConfig(); err != nil {
		return err
	}
	fmt.Println("Using config file:", viper.ConfigFileUsed())

	viper.SetConfigFile(getFilename(runMode))
	if err := viper.MergeInConfig(); err != nil {
		return err
	}
	fmt.Println("Merge config file:", viper.ConfigFileUsed())

	checkMissingResourceEnvVars()
	viper.SetEnvPrefix(envPrefix)
	viper.AutomaticEnv()
	return nil
}

func getFilename(runMode string) string {
	configFile := filepath.Join(defaultPath, defaultFileBaseName+"."+runMode+"."+defaultFileType)
	if _, err := os.Stat(configFile); err == nil {
		return configFile
	}
	fmt.Printf("WARN: create new config file %s\n", configFile)
	file, err := os.Create(configFile)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		return ""
	}
	file.Close()
	return configFile
}

// checkMissingResourceEnvVars will read the environment for equivalent config variables to set
func checkMissingResourceEnvVars() {
	//if !c.Resource.DaemonSet && os.Getenv("KW_DAEMONSET") == "true" {
	//	c.Resource.DaemonSet = true
	//}
}

func Save() error {
	return viper.WriteConfigAs(getFilename("curr"))
}

func Get(item string) interface{} {
	return viper.Get(item)
}

func GetString(item string) string {
	return viper.GetString(item)
}

func GetStringSlice(item string) []string {
	return viper.GetStringSlice(item)
}

func Set(key string, value interface{}) {
	viper.Set(key, value)
}

func GetInt(item string) int {
	return viper.GetInt(item)
}

func GetBool(item string) bool {
	return viper.GetBool(item)
}
