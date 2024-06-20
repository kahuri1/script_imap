package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func main() {
	if err := initConfig(); err != nil {
		logrus.Fatalf("error initializing configs: %s", err.Error())
	}

	c := Connect()
	defer c.Logout()
	lastIUD := ""
	from := uint32(1)
	for {

		mbox, err := c.Select("INBOX", false)
		if err != nil {
			log.Fatal(err)
		}

		// Get the last 4 messages // TODO если писем меньше то будет выводить последнее нужна проверка еще на уид
		to := mbox.Messages
		seqset := new(imap.SeqSet)
		seqset.AddRange(from, to)

		messages := make(chan *imap.Message)
		done := make(chan error, 1)
		go func() {
			done <- c.Fetch(seqset, []imap.FetchItem{imap.FetchEnvelope}, messages)
		}()

		for msg := range messages {
			if msg.Envelope.MessageId == lastIUD {
				continue
			}

			log.Println("* " + msg.Envelope.Subject)
			lastIUD = msg.Envelope.MessageId
			from++
			saveLastMessageInfo(int64(from), lastIUD)
			//получения вложения

			if err := <-done; err != nil {
				log.Fatal(err)
			}

			time.Sleep(time.Second * 5)
		}
	}
}

func saveLastMessageInfo(lastID int64, uid string) error {
	ms := &LastMessageInfo{
		CountMessage: lastID,
		LastUID:      uid,
	}

	file, err := os.Create("LastMessageInfo.json")
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	err = encoder.Encode(ms)
	if err != nil {
		return err
	}

	fmt.Println("LastMessageInfo saved successfully.")

	return nil
}

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
		Imap:     viper.GetString("imap"),
		Email:    viper.GetString("email"),
		Password: viper.GetString("password"),
	}
	return cfg

}

func Connect() *client.Client {
	log.Println("Connecting to server...")

	cfg := initAuth()

	// Connect to server
	c, err := client.DialTLS(cfg.Imap, nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected")

	// Login
	if err := c.Login(cfg.Email, cfg.Password); err != nil {
		log.Fatal(err)
	}
	log.Println("Logged in")
	return c
}
