package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
)

type Config struct {
	Imap     string
	Email    string
	Password string
}

func main() {
	log.Println("Connecting to server...")

	// Connect to server
	c, err := client.DialTLS("imap.mail.ru:993", nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected")

	// Don't forget to logout
	defer c.Logout()

	// Login
	if err := c.Login("bulanva@internet.ru", "QLQgCWrcCCwUz1RGdxqb"); err != nil {
		log.Fatal(err)
	}
	log.Println("Logged in")

	// Select INBOX
	//TODO INBOX - нужно вывести в конфиг

	//if mbox.Messages > 3 {
	//	// We're using unsigned integers here, only subtract if the result is > 0
	//	from = mbox.Messages - 3
	//}
	//log.Println("Last 4 messages:")
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

		time.Sleep(time.Second * 5)
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
