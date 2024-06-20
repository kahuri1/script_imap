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

	cfg := initAuth()

	c, err := ConnectServer(cfg)

	if err != nil {
		logrus.Fatalf("error connect server: %s", err.Error())
	}

	defer func() {
		if err := c.Logout(); err != nil {
			logrus.Errorf("error logging out: %s", err.Error())
		}
	}()

	err = loginToMail(cfg, c)
	if err != nil {
		logrus.Fatalf("error loginToMail: %s", err.Error())
	}

	lastIUD := ""
	from := uint32(1)
	for {

		mbox, err := c.Select("INBOX", false)
		if err != nil {
			log.Fatal(err)
		}
		//log.Println("Flags for INBOX:", mbox.Flags)

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
		}

		if err := <-done; err != nil {
			log.Fatal(err)
		}
		fmt.Println("Ожидание письма")
		time.Sleep(time.Second * 2)
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
	defer func() {
		if err := file.Close(); err != nil {
			logrus.Errorf("error logging out: %s", err.Error())
		}
	}()

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
