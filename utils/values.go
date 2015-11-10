package utils

import (
	"github.com/spf13/viper"
)


func init() {
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath("./config/")
	err := viper.ReadInConfig() 
	if err != nil {
		panic(err.Error())
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

func GetDBName() string {
	return viper.GetString("DB_NAME")
}

func GetUserDB() string {
	return viper.GetString("DB_USER")
}

func GetPassDB() string {
	return viper.GetString("DB_PASS")
}

func GetHTTPDir() string {
	return viper.GetString("HTTP_DIR")
}

func GetImgDir() string {
	return GetHTTPDir() + "images/"
}

func GetTempImgDir() string {
	return GetImgDir() + "temp/"
}

func ResetDB() bool {
	return viper.GetString("DB_RESET") == "yes"
}

func GetAdminPass() string {
	return viper.GetString("ADMIN_PASS")
}

func GetServerHost() string {
	return viper.GetString("SERVER_HOST")
}

func GetServerPort() string {
	return viper.GetString("SERVER_PORT")
}

func GetServerUrl() string {
	host := GetServerHost()
	port := GetServerPort()
	if port == "80" {
		return host
	} else {
		return host + ":" + port
	}	
}