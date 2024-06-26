package main

import (
	"bytes"
	"encoding/base64"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-message/mail"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"
)

func main() {

	logFile, err := os.OpenFile("errors.txt", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Println("Failed to open log file:", err)
		return
	}
	defer logFile.Close()

	log.SetOutput(logFile)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	if err := initConfig(); err != nil {
		log.Printf("error initializing configs: %s", err.Error())
	}
	cfg := initAuth()

	c, err := ConnectServer(cfg)
	if err != nil {
		log.Printf("error connect server: %s", err.Error())
	}

	err = loginToMail(cfg, c)
	if err != nil {
		log.Printf("error loginToMail: %s", err.Error())
	}
	defer func() {
		if err := c.Logout(); err != nil {
			log.Printf("error logging out: %s", err.Error())
		}
	}()

	from := cfg.From

	lastIUD := cfg.LastUID

	for {
		mbox, err := c.Select("INBOX", false)
		if err != nil {
			log.Println(err)
		}

		to := uint32(1029) //cfg.Messages
		//для первого запуска чтобы получить кол-во писем
		if from == 0 {
			from = mbox.Messages
		}
		if from > to {
			return
		}
		seqset := new(imap.SeqSet)
		sect := &imap.BodySectionName{}
		seqset.AddRange(from, to)

		messages := make(chan *imap.Message, 1)
		done := make(chan error, 1)
		section := &imap.BodySectionName{}

		go func() {
			done <- c.Fetch(seqset, []imap.FetchItem{sect.FetchItem(), imap.FetchEnvelope, imap.FetchBodyStructure}, messages)
		}()

		for msg := range messages {
			if msg.Envelope.MessageId == lastIUD {
				continue
			}

			//params := msg.BodyStructure.Params
			//charset := params["charset"]

			mr, err := mail.CreateReader(msg.GetBody(section))
			if err != nil {
				log.Fatal(err)
			}

			for {
				p, err := mr.NextPart()
				if err == io.EOF {
					break
				} else if err != nil {
					continue
				}

				err = processPart(p, cfg, msg)
				if err != nil {
					log.Println(err)
				}
			}

			lastIUD = msg.Envelope.MessageId
			from++
			err = SetDefaultValue(from, lastIUD)
			if err != nil {
				log.Fatalf("error setting default from: %s", err.Error())
			}
		}
		if err := <-done; err != nil {
			log.Fatal(err)
		}

		time.Sleep(time.Second * 5)

	}
}

func convertToUTF8(charset string, data []byte) ([]byte, error) {
	if charset == "windows-1251" {
		utf8Reader := transform.NewReader(bytes.NewReader(data), charmap.Windows1251.NewDecoder())
		utf8Data, err := ioutil.ReadAll(utf8Reader)
		if err != nil {
			return nil, err
		}
		return utf8Data, nil
	}
	if charset == "koi8-r" {
		utf8Reader := transform.NewReader(bytes.NewReader(data), charmap.KOI8R.NewDecoder())
		utf8Data, err := ioutil.ReadAll(utf8Reader)
		if err != nil {
			return nil, err
		}
		return utf8Data, nil
	}
	// Add more cases for different charsets if needed
	return data, nil
}

func processPart(p *mail.Part, cfg Config, msg *imap.Message) error {
	switch h := p.Header.(type) {
	case *mail.AttachmentHeader:
		filename, _ := h.Filename()
		if strings.Contains(filename, "windows-1251") {
			start := len("=?windows-1251?B?")
			end := len(filename) - len("?=")
			filename = filename[start:end]
			decoded, err := base64.StdEncoding.DecodeString(filename)
			if err != nil {
				log.Println("Ошибка декодирования:", err)
			}
			filename = string(decoded)
		}
		if strings.Contains(filename, "Cp1251") {
			start := len("=?Cp1251?B?")
			end := len(filename) - len("?=")
			filename = filename[start:end]
			decoded, err := base64.StdEncoding.DecodeString(filename)
			decoder := charmap.Windows1251.NewDecoder()
			filenameDecoder, err := decoder.String(string(decoded))
			if err != nil {
				log.Println("Ошибка декодирования:", err)
			}
			filename = string(filenameDecoder)
		}

		if strings.Contains(filename, "koi8-r") {
			start := len("=?windows-1251?B?")
			end := len(filename) - len("?=")
			filename = filename[start:end]
			decoded, err := base64.StdEncoding.DecodeString(filename)
			if err != nil {
				log.Println("Ошибка декодирования:", err)
			}
			filename = string(decoded)
		}
		b, _ := ioutil.ReadAll(p.Body)
		err := ioutil.WriteFile(cfg.Storage+filename, b, 0777)

		if err != nil {
			log.Printf("От:%s , заголовок: %S, имя вложения: %s, кодировка:%s", msg.Envelope.MessageId, msg.Envelope.From, msg.Envelope.Subject, msg.BodyStructure.Params)
			return err
		}
	}
	return nil
}
