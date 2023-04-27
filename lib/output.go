package lib

import (
	"encoding/json"
	"github.com/slack-go/slack"
	"log"
	"os"
)

func SaveMessagesAsJSONFile(messages []slack.Message, fileName string) error {
	file, err := os.Create(fileName)
	if err != nil {
		log.Fatal("Create file error: ", err)
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Fatal("Close file error: ", err)
		}
	}(file)

	data, err := json.MarshalIndent(messages, "", "  ")
	if err != nil {
		log.Fatal("JSON Marshal error: ", err)
		return err
	}

	_, err = file.Write(data)
	if err != nil {
		log.Fatal("Write to file error: ", err)
		return err
	}

	return nil
}
