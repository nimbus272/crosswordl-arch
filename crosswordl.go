package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	DATE_FORMAT  = "2006-01-02" // YYYY-MM-DD
	DOWNLOAD_URL = "https://www.onlinecrosswords.net/printable-daily-crosswords-1.pdf"
)

func main() {
	crosswordPath, currentDate := initialization()
	deleteOldFiles(crosswordPath)
	crossword := downloadCrossword(currentDate, crosswordPath)
	sendNotification(crossword)
}

func initialization() (string, string) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("Failed to retrieve user home directory with error: ", err)
	}
	crosswordPath := homeDir + "/Documents/crosswords"
	t := time.Now()
	currentDate := t.Format(DATE_FORMAT)
	mkDirErr := os.Mkdir(crosswordPath, 0700)
	if mkDirErr != nil {
		log.Print("Crossword Directory already exists, skipping creation..")
	}
	return crosswordPath, currentDate
}

func sendNotification(crossword string) {
	output, err := exec.Command("notify-send", "--action=Yes", "--action=No", "Crosswordl", "New Crossword Downloaded! Would you like to open it?").Output()
	if err != nil {
		log.Fatal("Failed to send notification with error: ", err)
	}
	outputString := string(output[:])
	if strings.Contains(outputString, "0") {
		log.Print(crossword)
		_, err := exec.Command("xdg-open", crossword).Output()
		if err != nil {
			log.Fatal("Failed to start pdf app with error: ", err)
		}
	}
	if strings.Compare(outputString, "0") == 0 {
		log.Print(crossword)
		_, err := exec.Command("xdg-open", crossword).Output()
		if err != nil {
			log.Fatal("Failed to start pdf app with error: ", err)
		}
	}
}

func deleteOldFiles(crosswordPath string) {
	files, err := os.ReadDir(crosswordPath)
	if err != nil {
		log.Fatal("Failed To Read Crossword Directory with error: ", err.Error())
	}

	for _, file := range files {

		fileName := file.Name()
		splitString := strings.Split(fileName, ".")
		name := splitString[0]
		fileDate, dateParseErr := time.Parse(DATE_FORMAT, name)
		if dateParseErr != nil {
			log.Fatal("Failed to parse date from filename "+name+" with error: ", dateParseErr)
		}
		aWeekAgo := time.Now().AddDate(0, 0, -8)
		if fileDate.Before(aWeekAgo) {
			err := os.Remove(crosswordPath + "/" + fileName)
			if err != nil {
				log.Fatal("Failed to delete old crossword with error: ", err)
			}

		}
	}
}

func downloadCrossword(currentDate string, crosswordPath string) string {
	// Create blank file
	crossword := crosswordPath + "/" + currentDate
	file, err := os.Create(crossword)
	if err != nil {
		log.Fatal("Failed to create file with error: ", err)
	}

	client := http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			req.URL.Opaque = req.URL.Path
			return nil
		},
	}

	resp, err := client.Get(DOWNLOAD_URL)
	if err != nil {
		log.Fatal("Failed to download Crossword with error: ", err)
	}

	defer resp.Body.Close()

	_, writeErr := io.Copy(file, resp.Body)
	if writeErr != nil {
		log.Fatal("Failed to write to file from response with error: ", writeErr)
	}

	defer file.Close()
	log.Print("Crossword downloaded successfully")
	return crossword
}
