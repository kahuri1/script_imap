package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-message/mail"
	"github.com/sirupsen/logrus"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
	"io"
	"io/ioutil"
	"log"
	"net/textproto"
	"strings"
	"time"
)

func main() {
	//log := logrus.New()
	//file, err := os.OpenFile("logfile.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	//if err != nil {
	//	log.Fatal("Невозможно открыть файл логов: ", err)
	//}
	//log.SetOutput(file)
	//defer file.Close()
	//// Сохранение ошибки без завершения программы

	if err := initConfig(); err != nil {
		log.WithError(err).Error("error initializing configs: %s", err.Error())
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

		to := uint32(1029) //cfg.Messages
		//для первого запуска чтобы получить кол-во писем
		if from == 0 {
			from = mbox.Messages
		}
		if from == to {
			break
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

			params := msg.BodyStructure.Params
			charset := params["charset"]
			fmt.Println(charset)

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

				err = processPart(p, cfg)
				if err != nil {
					log.Println(err)
				}
			}

			lastIUD = msg.Envelope.MessageId
			from++
			err = SetDefaultValue(from, lastIUD)
			if err != nil {
				logrus.Fatalf("error setting default from: %s", err.Error())
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
	// Add more cases for different charsets if needed
	return data, nil
}

func processPart(p *mail.Part, cfg Config) error {
	switch h := p.Header.(type) {
	case *textproto.MIMEHeader:
		contentType := h.Get("Content-Type")
		params := strings.Split(contentType, ";")
		charset := "UTF-8" // Assume UTF-8 if charset is not specified

		for _, param := range params {
			if strings.Contains(param, "charset") {
				charset = strings.Split(param, "=")[1]
				break
			}
		}

		b, err := ioutil.ReadAll(p.Body)
		if err != nil {
			return err
		}

		utf8Data, err := convertToUTF8(charset, b)
		if err != nil {
			return err
		}
		fmt.Println(string(utf8Data))
	case *mail.AttachmentHeader:
		filename, _ := h.Filename()
		if strings.Contains(filename, "windows-1251") {
			fmt.Println(filename)
			start := len("=?windows-1251?B?")
			end := len(filename) - len("?=")
			filename = filename[start:end]
			//filename = "=" + filename + "="
			fmt.Println(filename)
			decoded, err := base64.StdEncoding.DecodeString(filename)
			if err != nil {
				fmt.Println("Ошибка декодирования:", err)
			}

			fmt.Println(string(decoded))
			fmt.Println(filename)
			filename = string(decoded)
		}

		if strings.Contains(filename, "koi8-r") {
			filename = "koi8-r_test.csv"
		}
		b, _ := ioutil.ReadAll(p.Body)
		err := ioutil.WriteFile(cfg.Storage+filename, b, 0777)

		if err != nil {
			return err
		}
	}
	return nil
}
