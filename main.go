package main

import (
	"github.com/emersion/go-imap"
	"github.com/sirupsen/logrus"
	"log"
	"time"
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

	err = loginToMail(cfg, c)
	if err != nil {
		logrus.Fatalf("error loginToMail: %s", err.Error())
	}
	defer func() {
		if err := c.Logout(); err != nil {
			logrus.Errorf("error logging out: %s", err.Error())
		}
	}()

	from := cfg.From
	lastIUD := cfg.LastUID

	for {
		mbox, err := c.Select("INBOX", false)
		if err != nil {
			log.Fatal(err)
		}
		//log.Println("Flags for INBOX:", mbox.Flags)
		// Get the last 4 messages
		// Get the last 4 messages // TODO если писем меньше то будет выводить последнее нужна проверка еще на уид
		to := mbox.Messages
		seqset := new(imap.SeqSet)
		seqset.AddRange(from, to)

		messages := make(chan *imap.Message, 1)
		done := make(chan error, 1)
		go func() {
			done <- c.Fetch(seqset, []imap.FetchItem{imap.FetchEnvelope}, messages)
		}()

		for msg := range messages {
			if msg.Envelope.MessageId == lastIUD {
				continue
			} else {
				log.Println("* " + msg.Envelope.Subject)

				lastIUD = msg.Envelope.MessageId
				from++

				err := SetDefaultFrom(from)
				if err != nil {
					logrus.Fatalf("error setting default from: %s", err.Error())
				}
			}
		}

		if err := <-done; err != nil {
			log.Fatal(err)
		}

		time.Sleep(time.Second * 5)
	}
}
