package logit

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/brianbroderick/lantern/pkg/sql/projectpath"
)

func Append(filePrefix, str string) {
	now := time.Now()
	date := now.Format("2006-01-02")
	logFile := filepath.Join(projectpath.Root, "logs", fmt.Sprintf("%s-%s.log", filePrefix, date))
	// fmt.Println("Logging to: ", logFile)

	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	defer f.Close()

	logger := log.New(f, "", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
	logger.Println(str)
}
