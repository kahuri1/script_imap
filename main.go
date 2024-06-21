package main

import (
	"github.com/emersion/go-imap"
	"github.com/sirupsen/logrus"
	"io/ioutil"
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
		sect := &imap.BodySectionName{}
		seqset.AddRange(from, to)

		messages := make(chan *imap.Message, 1)
		done := make(chan error, 1)
		go func() {
			done <- c.Fetch(seqset, []imap.FetchItem{sect.FetchItem(), imap.FetchEnvelope}, messages)
		}()

		for msg := range messages {
			if msg.Envelope.MessageId == lastIUD {
				continue
			} else {
				log.Println("* " + msg.Envelope.Subject)
				for _, p := range msg.Body {

					data, err := ioutil.ReadAll(p) // Читаем содержимое файла в []byte
					if err != nil {
						log.Fatal(err)
					}

					//TODO тут нам нужно разобраться с расширением
					err = ioutil.WriteFile("output.html", data, 0644)
					if err != nil {
						log.Fatal(err)
					}

				}

				lastIUD = msg.Envelope.MessageId
				from++

				err := SetDefaultValue(from, lastIUD)
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
