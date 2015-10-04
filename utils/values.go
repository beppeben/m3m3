package utils

import (
	"github.com/spf13/viper"
	"log"
)

var (
	def_DB_RESET		string = "no"
	def_IMG_DIR 		string = "/var/www/m3m3/images/"
)

func init() {

	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath("./config/")
	
	viper.SetDefault("DB_RESET", def_DB_RESET)
	viper.SetDefault("IMG_DIR", def_IMG_DIR)

	err := viper.ReadInConfig() 

	if err != nil { 
		log.Printf("[OMG] Cannot read config file : %s", err)		
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

func GetImgDir() string {
	return viper.GetString("IMG_DIR")
}

func ResetDB() bool {
	return viper.GetString("DB_RESET") == "yes"
}
