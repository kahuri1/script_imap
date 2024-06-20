package main

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
)

func saveLastMessageInfo(lastID int64, uid string) error {
	ms := &LastMessageInfo{
		CountMessage: lastID,
		LastUID:      uid,
	}
	err := SetDefaultUID(uid)
	if err != nil {
		logrus.Errorf("error save uid: %s", err.Error())
	}
	file, err := os.Create("LastMessageInfo.json")
	if err != nil {
		return err
	}
	defer func() {
		if err := file.Close(); err != nil {
			logrus.Errorf("error logging out: %s", err.Error())
		}
	}()

	encoder := json.NewEncoder(file)
	err = encoder.Encode(ms)
	if err != nil {
		return err
	}

	fmt.Println("LastMessageInfo saved successfully.")

	return nil
}
