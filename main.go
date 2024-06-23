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
		section := &imap.BodySectionName{
			/* BodyPartName: imap.BodyPartName{
				Specifier: imap.MIMESpecifier,
				Fields:    []string{"X-Priority", "Priority"},
			}, */
		}

		go func() {
			done <- c.Fetch(seqset, []imap.FetchItem{sect.FetchItem(), imap.FetchEnvelope}, messages)
		}()

		for msg := range messages {
			if msg.Envelope.MessageId == lastIUD {
				continue
			} else {
				log.Println("* " + msg.Envelope.Subject)

				mr, err := mail.CreateReader(msg.GetBody(section))
				if err != nil {
					log.Fatal(err)
				}

				// Print some info about the message
				header := mr.Header
				if date, err := header.Date(); err == nil {
					log.Println("Date:", date)
				}
				if from, err := header.AddressList("From"); err == nil {
					log.Println("From:", from)
				}
				if to, err := header.AddressList("To"); err == nil {
					log.Println("To:", to)
				}

				if prio := header.Get("X-Priority"); prio != "" {
					log.Println("Priority:", prio)
				}

				if subject, err := header.Subject(); err == nil {
					log.Println("Subject:", subject)
				}

				if conxtdis := header.Get("Content-Disposition"); conxtdis != "" {
					log.Println("Content-Disposition:", conxtdis)
				}

				// Process each message's part
				for {
					p, err := mr.NextPart()
					if err == io.EOF {
						break
					} else if err != nil {
						log.Fatal(err)
					}

					disp := p.Header.Get("Content-Disposition")

					switch h := p.Header.(type) {
					case *mail.InlineHeader:
						// This is the message's text (can be plain-text or HTML)
						b, _ := ioutil.ReadAll(p.Body)
						if disp != "" {

							contentID := h.Get("Content-ID")

							_, pr, _ := h.ContentType()

							filename := fmt.Sprintf("%s.%s", contentID, pr["name"])

							ioutil.WriteFile(filename, b, 0777)

						} else {
							log.Printf("Got ====: %s", string(b))
						}
					case *mail.AttachmentHeader:

						log.Printf("Got attachment==========")
						// This is an attachment
						filename, _ := h.Filename()

						log.Printf("Got attachment: %v", filename)
						b, errp := ioutil.ReadAll(p.Body)
						fmt.Println("errp ===== :", errp)
						err := ioutil.WriteFile(filename, b, 0777)

						if err != nil {
							log.Println("attachment err: ", err)
						}
					}
				}

			}
		}
		if err := <-done; err != nil {
			log.Fatal(err)
		}

		time.Sleep(time.Second * 5)

	}
}
