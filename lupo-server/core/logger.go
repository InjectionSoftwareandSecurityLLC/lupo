package core

import (
	"log"
	"os"
)

// LogData - wrapper function to use golang's built in logger and append all operational data to a central log file
func LogData(data string) error {
	file, err := os.OpenFile(".lupo.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	defer file.Close()

	log.SetOutput(file)
	log.Print(data)

	return nil
}

func ChatLog(data string) error {
	file, err := os.OpenFile(".lupo.chat.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	defer file.Close()

	log.SetOutput(file)
	log.Print(data)

	return nil
}
