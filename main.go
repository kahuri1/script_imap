package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/emersion/go-imap"
)

type Config struct {
	Imap     string
	Email    string
	Password string
}

func main() {

	c := Connect(Config{})

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
