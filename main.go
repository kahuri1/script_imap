package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/spf13/viper"
)

type Config struct {
	Imap     string
	Email    string
	Password string
}

func main() {

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
	c, err := client.DialTLS(cfg.Imap, nil)
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
			for msg := range messages {
				if msg.BodyStructure == nil {
					log.Println("Сообщение не содержит вложений")
					continue
				}
				entity, err := c.ReadTo(r)
				if err != nil {
					log.Fatal(err)
				}
				multiPartReader := entity.MultipartReader()
				e, err := multiPartReader.NextPart()
                        if err == io.EOF {
                            break
                        }
				data, err := io.ReadAll(e.Body)
                            if err != nil {
                                log.Fatal(err)
                            }
				for _, part := range msg.Body {
								fileName := "/tmp/" + params["name"]
								if err := os.WriteFile(fileName, data, 0644); err != nil {
									log.Fatal(err)
								}
	
								log.Printf("Файл %s сохранен", fileName)
							}
						
					
				

		if err := <-done; err != nil {
			log.Fatal(err)
		}

		time.Sleep(time.Second * 5)
	}
}
	}
}

type LastMessageInfo struct {
	CountMessage int64  `json:"count_message"`
	LastUID      string `json:"last_uid"`
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
