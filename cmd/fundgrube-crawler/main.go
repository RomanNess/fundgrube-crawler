package main

import (
	"fmt"
	"fundgrube-crawler/alert"
	"fundgrube-crawler/crawler"
	log "github.com/sirupsen/logrus"
	easy "github.com/t-tomalak/logrus-easy-formatter"
	"os"
	"time"
)

var LOG_FILE = fmt.Sprintf("/tmp/fundgrube-%s.txt", time.Now().Format("2006-01-02T15-04-05"))

func main() {
	configureLogger()
	if envBool("LOG_TO_FILE") {
		defer mailAlertOnPanic()
	}

	if !envBool("SKIP_CRAWLING") {
		err := crawler.CrawlPostings(envBool("MOCKED_POSTINGS"))
		if err != nil {
			panic(err)
		}
	}

	crawler.SearchDeals()
}

func configureLogger() {
	log.SetFormatter(&easy.Formatter{
		TimestampFormat: "2006-01-02T15:04:05Z07",
		LogFormat:       "%time% [%lvl%] %msg%\n",
	})

	level, err := log.ParseLevel(env("LOG_LEVEL", "info"))
	if err != nil {
		panic(err)
	}
	log.SetLevel(level)

	if envBool("LOG_TO_FILE") {
		file, err := os.OpenFile(LOG_FILE, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			panic(err)
		}

		log.Infof("Logging into '%s'", LOG_FILE)
		log.SetOutput(file)
	}
}

func mailAlertOnPanic() {
	if r := recover(); r != nil {
		var errorString string
		switch x := r.(type) {
		case string:
			errorString = x
		case error:
			errorString = x.Error()
		default:
			errorString = "Unknown panic"
		}

		subject := fmt.Sprint("💥Panic occurred : ", errorString)
		contentBytes := getContentBytes()
		err := alert.SendAlertMailBytes(subject, contentBytes)
		if err != nil {
			log.Fatalf("Failed to alert abount panic '%s' via mail. Send error '%s'", r.(string), err.Error())
		}
		log.Errorln("💥Panic occurred. Send alert mail.", r)
	}
}

func getContentBytes() []byte {
	if envBool("LOG_TO_FILE") {
		contentBytes := []byte("\n\nLogs:\n\n")
		logBytes, err := os.ReadFile(LOG_FILE)
		if err != nil {
			panic(err)
		}
		return append(contentBytes, logBytes...)
	}
	return nil
}

func envBool(key string) bool {
	return os.Getenv(key) == "true"
}

func env(key string, defaultValue string) string {
	value, present := os.LookupEnv(key)
	if present {
		return value
	}
	return defaultValue
}
