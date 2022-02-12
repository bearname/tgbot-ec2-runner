package common

import (
	"flag"
	"github.com/joho/godotenv"
	"github.com/sirupsen/log"
	"os"
)

func LoadEnvFileIfNeeded(envFilename string) {
	var isNeedLoadEnvFile string
	flag.StringVar(&isNeedLoadEnvFile, "d", "false", "is need load .env file")
	flag.Parse()
	if isNeedLoadEnvFile == "true" {
		err := godotenv.Load(envFilename)
		if err != nil {
			log.Fatal("Error loading .env file")
		}
	}
}

func LogToFileIfNeeded(filename string) {
	log.SetFormatter(&log.JSONFormatter{})
	var isNeedFileLog string
	flag.StringVar(&isNeedFileLog, "log", "false", "is need load .env file")
	flag.Parse()
	if isNeedFileLog == "false" {
		file, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 6444)
		if err == nil {
			log.SetOutput(file)
			defer func(file *os.File) {
				err = file.Close()
				if err != nil {
					log.Error(err)
				}
			}(file)
		}
	}
}
