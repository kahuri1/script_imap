package main

import (
	"fmt"
	"github.com/emersion/go-imap/client"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"log"
)

func initConfig() error {

	viper.AddConfigPath(".")
	viper.SetConfigName("config")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("файл config не найден: %w", err))
	}
	return viper.ReadInConfig()
}

func initAuth() Config {
	cfg := Config{
		Imap:      viper.GetString("imap"),
		Email:     viper.GetString("email"),
		Password:  viper.GetString("password"),
		LastUID:   viper.GetString("last_uid"),
		From:      viper.GetUint32("from"),
		Storage:   viper.GetString("file_storage_directory"),
		logPath:   viper.GetString("file_log_path"),
		timeDelay: viper.GetUint32("time_delay"),
	}

	return cfg

}

func ConnectServer(cfg Config) (*client.Client, error) {
	log.Println("Connecting to server...")
	// Connect to server
	c, err := client.DialTLS(cfg.Imap, nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected")
	return c, nil
}

func loginToMail(cfg Config, c *client.Client) error {
	// Login
	if err := c.Login(cfg.Email, cfg.Password); err != nil {
		log.Fatal(err)
	}
	log.Println("Logged in")

	return nil
}

func SetDefaultUID(LastUID string) error {
	viper.Set("lastuid", LastUID)
	if err := viper.WriteConfig(); err != nil {
		logrus.Fatalf("error saving LastUID to config: %s", err.Error())
	}
	return nil
}

func SetDefaultValue(from uint32, uid string) error {
	viper.Set("from", from)
	if err := viper.WriteConfig(); err != nil {
		logrus.Fatalf("error saving from to config: %s", err.Error())
	}

	viper.Set("last_uid", uid)
	if err := viper.WriteConfig(); err != nil {
		logrus.Fatalf("error saving from to config: %s", err.Error())
	}

	return nil
}
