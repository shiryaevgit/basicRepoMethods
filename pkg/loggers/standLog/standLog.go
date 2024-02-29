package standLog

import (
	"log"
	"os"
)

func LoadStandLog(filePath string) (*os.File, error) {
	fileLog, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		log.Printf("openFile(error.log): %v", err)
	}

	log.SetOutput(fileLog)
	log.SetFlags(log.Lshortfile | log.Ltime)

	return fileLog, nil
}
