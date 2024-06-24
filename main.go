package main

import (
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

		to := mbox.Messages
		seqset := new(imap.SeqSet)
		sect := &imap.BodySectionName{}
		seqset.AddRange(from, to)

		messages := make(chan *imap.Message, 1)
		done := make(chan error, 1)
		section := &imap.BodySectionName{}

		go func() {
			done <- c.Fetch(seqset, []imap.FetchItem{sect.FetchItem(), imap.FetchEnvelope}, messages)
		}()

		for msg := range messages {
			if msg.Envelope.MessageId == lastIUD {
				continue
			} else {

				mr, err := mail.CreateReader(msg.GetBody(section))
				if err != nil {
					log.Fatal(err)
				}

				for {
					p, err := mr.NextPart()
					if err == io.EOF {
						break
					} else if err != nil {
						log.Fatal(err)
					}

					switch h := p.Header.(type) {

					case *mail.AttachmentHeader:
						filename, _ := h.Filename()

						b, _ := ioutil.ReadAll(p.Body)
						err := ioutil.WriteFile(cfg.Storage+filename, b, 0777)

						if err != nil {
							break
						}
					}
				}

				lastIUD = msg.Envelope.MessageId
				from++
				err = SetDefaultValue(from, lastIUD)
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
