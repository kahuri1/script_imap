package main

import (
	"fmt"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-message/mail"
	"github.com/sirupsen/logrus"
	"io"
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

	defer func() {
		if err := c.Logout(); err != nil {
			logrus.Errorf("error logging out: %s", err.Error())
		}
	}()

	err = loginToMail(cfg, c)
	if err != nil {
		logrus.Fatalf("error loginToMail: %s", err.Error())
	}

	lastIUD := cfg.LastUID
	from := uint32(1)
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

		messages := make(chan *imap.Message)
		done := make(chan error, 1)

		 section := &imap.BodySectionName{}
		items := [] imap. FetchItem{section. FetchItem()}
		go func() {
			done <- c.Fetch(seqset, items, messages)
		}()

		// тут нужно поменять или вообще убрать рейнд, чтобы он не считывал письма, у нас же есть последний UID
		for msg := range messages {
			if msg.Envelope.MessageId == lastIUD {
				continue
			} else {

				log.Println("* " + msg.Envelope.Subject)
				lastIUD = msg.Envelope.MessageId
				from++
				//вложение
				r := msg. GetBody(section)
				if r == nil {
					log. Fatal("Server didn't returned message body")
				}
				if err := <-done; err != nil {
					log.Fatal(err)
				}

				m, err := mail.ReadMessage(r)
				if err != nil {
					log.Fatal(err)
				}

				header := m.Header
				log.Println("Date:", header.Get("Date"))
				log.Println("From:", header.Get("From"))
				log.Println("To:", header.Get("To"))
				log.Println("Subject:", header.Get("Subject"))

				body, err := io.ReadAll(m.Body)
				if err != nil {
					log.Fatal(err)
				}
				log.Println(body)

					//вложение

					err = saveLastMessageInfo(int64(from), lastIUD)
					if err != nil {
						logrus.Errorf("error saving last message to file: %s", err.Error())
					}
				}
			}

			if err := <-done; err != nil {
				log.Fatal(err)
			}
			fmt.Println("Ожидание письма")
			time.Sleep(time.Second * 2)
		}
	}
}
