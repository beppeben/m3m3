package utils

import (
	"github.com/spf13/viper"
	"log"
)

var (
	def_DB_RESET string = "no"
)

func init() {

	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath("./config/")

	err := viper.ReadInConfig() // Find and read the config file

	if err != nil { // Handle errors reading the config file
		log.Printf("[OMG] Cannot read config file : %s", err)
		viper.SetDefault("DB_RESET", def_DB_RESET)
	}
}

func GetServiceEmail() string {
	return viper.GetString("EMAIL")
}

func GetEmailPass() string {
	return viper.GetString("EMAIL_PASS")
}

func GetSMTP() string {
	return viper.GetString("SMTP")
}

func GetSMTPPort() string {
	return viper.GetString("SMTP_PORT")
}

func GetUserDB() string {
	return viper.GetString("DB_USER")
}

func GetPassDB() string {
	return viper.GetString("DB_PASS")
}

func ResetDB() bool {
	return viper.GetString("DB_RESET") == "yes"
}
