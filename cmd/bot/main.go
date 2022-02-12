package main

import (
	"aws/cmd/bot/config"
	"aws/internal/bot/application/ec2ser"
	"aws/internal/common"
	"bytes"
	"fmt"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
	"strings"
)

func main() {
	common.LogToFileIfNeeded(".env")
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	tgToken := os.Getenv("TG_TOKEN")

	if len(tgToken) == 0 {
		log.Fatal("TG_TOKEN not setted")
	}
	conf := &config.Config{
		DbAddress:      "ec2-34-228-100-83.compute-1.amazonaws.com",
		DbName:         "d7e2b570mkmiab",
		DbUser:         "rgmfhxwictxgag",
		DbPassword:     "8dec647b9149246002a753bf59d841adb6be5e54dd1e98afa716e31e3b7c83cc",
		MaxConnections: 10,
		AcquireTimeout: 1,
	}

	connector, err := config.GetConnector(conf)
	if err != nil {
		log.Fatal(err.Error())
	}

	pool, err := config.NewConnectionPool(connector)
	if err != nil {
		log.Fatal(err.Error())
	}
	awsSess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	service := ec2ser.NewService(pool, awsSess)
	menu := NewMenu(tgToken, service)
	err = menu.handle()
	if err != nil {
		log.Println(err)
	}
}

type Menu struct {
	bot        *tgbotapi.BotAPI
	ec2Service ec2ser.Service
}

func NewMenu(tgToken string, service *ec2ser.Service) *Menu {
	bot, err := tgbotapi.NewBotAPI(tgToken)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	m := new(Menu)
	m.bot = bot
	m.ec2Service = *service
	return m
}

func (m *Menu) handle() error {
	u := tgbotapi.NewUpdate(0)
	updates, err := m.bot.GetUpdatesChan(u)
	if err != nil {
		return err
	}
	for update := range updates {
		m.handleMessage(update)
	}

	return nil
}

func (m *Menu) handleMessage(update tgbotapi.Update) {
	message := update.Message
	if message == nil {
		return
	}

	if !message.IsCommand() {
		return
	}

	text := message.Text

	fmt.Println(text)
	command := message.Command()
	chatID := message.Chat.ID
	if command == "start" {
		msg := tgbotapi.NewMessage(chatID, "Hi. Available commands '/list', 'rdp'")
		msg.ReplyToMessageID = message.MessageID
		m.bot.Send(msg)
		return
	}

	fmt.Println("command")
	fmt.Println(command)

	if command == "list" {
		instancies := m.ec2Service.GetAvailableInstances()
		if len(instancies) == 0 {
			m.send(chatID, "Failed get aws ec2 instances ")
			return
		}
		response := ""
		for _, i := range instancies {
			state := i.State
			response += "Id " + i.Id + " URL " + i.PublicDnsName + " State " + *state.Name + "\n"
		}

		m.send(chatID, response)
		return
	}

	if command == "rdp" {
		split := strings.Split(message.Text, " ")
		m.send(chatID, message.Text+" "+strconv.Itoa(len(split)))

		if len(split) != 2 {
			m.send(chatID, "Invalid argument count. Usage /rdp <ec2ser instance name>")
			return
		}

		instanceId := split[1]
		rdpFile := m.ec2Service.GetRdpFile(instanceId)
		if len(rdpFile) == 0 {
			m.send(chatID, "Failed get rdp file for ec2 instance "+instanceId)
			return
		}

		buffer := bytes.Buffer{}
		buffer.Write([]byte(rdpFile))
		reader := tgbotapi.FileReader{Name: instanceId + ".rdp", Reader: &buffer, Size: -1}
		msg := tgbotapi.NewDocumentUpload(chatID, reader)
		send, err := m.bot.Send(msg)
		if err != nil {
			m.send(chatID, "Failed send file ")
			log.Fatal(err)
			return
		}
		fmt.Println(send)
		return
	}
	m.send(chatID, "Unknown command")

	fmt.Println(" fmt.Println(command)")
	fmt.Println(command)
}

func (m *Menu) send(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	m.bot.Send(msg)
}
