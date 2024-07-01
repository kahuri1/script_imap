package main

import (
	"github.com/emersion/go-imap"
	_ "github.com/emersion/go-message/charset"
	"github.com/emersion/go-message/mail"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"log"
	"os"
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

	logFile, err := os.OpenFile(cfg.logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Println("Failed to open log file:", err)
		return
	}
	defer logFile.Close()

	log.SetOutput(logFile)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

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

		if from == 0 {
			from = mbox.Messages
		}
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
				email := msg.Envelope.From[0].PersonalName
				if email == "" {
					email = msg.Envelope.From[0].MailboxName
				}
				mr, err := mail.CreateReader(msg.GetBody(section))
				if err != nil {
					log.Println(err)
				}

				for {
					p, err := mr.NextPart()
					if err == io.EOF {
						break
					} else if err != nil {
						log.Println(err, from)
					}

					switch h := p.Header.(type) {

					case *mail.AttachmentHeader:
						filename, _ := h.Filename()

						b, _ := ioutil.ReadAll(p.Body)
						err = ioutil.WriteFile(cfg.Storage+filename, b, 0777)

						if err != nil {
							log.Printf("Вложение не сохранено, имя файла: %s, почта: %s, номер письма: %d ", filename, email, from)
						}

					}
				}

				lastIUD = msg.Envelope.MessageId

				if from < to {
					from++
				}

				err = SetDefaultValue(from, lastIUD)
				if err != nil {
					log.Printf("error setting default from: %s", err.Error())
				}
			}
		}
		if err := <-done; err != nil {
			log.Fatal(err)
		}

		time.Sleep(time.Second * time.Duration(cfg.timeDelay))

	}
}
