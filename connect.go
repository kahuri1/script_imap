package main

import (
	"fmt"
	"log"

	"github.com/emersion/go-imap/client"
	"github.com/spf13/viper"
)

func Connect(Config) (c *client.Client) {
	log.Println("Connecting to server...")

	viper.AddConfigPath(".")
	viper.SetConfigName("config")
	err := viper.ReadInConfig()
	if err != nil {

		fmt.Println("error")
	}

	cfg := Config{
		Imap:     viper.GetString("imap"),
		Email:    viper.GetString("email"),
		Password: viper.GetString("password"),
	}

	// Connect to server
	c, err = client.DialTLS(cfg.Imap, nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected")

	// Don't forget to logout
	defer c.Logout()

	// Login
	if err := c.Login(cfg.Email, cfg.Password); err != nil {
		log.Fatal(err)
	}
	log.Println("Logged in")
	return c
}
