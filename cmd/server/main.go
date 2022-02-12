package main

import (
	"aws/cmd/bot/config"
	"aws/internal/common"
	"aws/internal/server/application/aws"
	"aws/internal/server/domain/task"
	"aws/internal/server/infrastructure/transport"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/jasonlvhit/gocron"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	common.LoadEnvFileIfNeeded(".env.backend")
	common.LogToFileIfNeeded("short.log")

	dbAddr := os.Getenv("DB_ADDR")
	if len(dbAddr) == 0 {
		log.Fatal("DB_ADDR not setted")
	}
	dbName := os.Getenv("DB_NAME")
	if len(dbName) == 0 {
		log.Fatal("DB_NAME not setted")
	}
	dbUser := os.Getenv("DB_USER")
	if len(dbUser) == 0 {
		log.Fatal("DB_USER not setted")
	}
	dbPass := os.Getenv("DB_PASS")
	if len(dbUser) == 0 {
		log.Fatal("DB_PASS not setted")
	}

	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8090"
	}

	conf := &config.Config{
		DbAddress:      dbAddr,
		DbName:         dbName,
		DbUser:         dbUser,
		DbPassword:     dbPass,
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

	go func() {
		<-gocron.Start()
	}()
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	svc := ec2.New(sess)
	repo := task.NewTaskRepo(pool)
	ec2Service := aws.NewEC2Service(svc)
	tasks, err := repo.Find()
	if err != nil {
		log.Println(err)
		return
	}
	for _, item := range tasks {
		log.Println(item)
		if !cronHandle(item.Start, ec2Service, item.Id, true) {
		}
		if !cronHandle(item.Stop, ec2Service, item.Id, false) {
		}
	}

	controller := transport.NewTaskController(repo, ec2Service)
	http.HandleFunc("/addTask", controller.AddTask)
	log.Println("Start")
	http.ListenAndServe(":"+port, nil)
	log.Println("End")

	graceFullShutdown()
}

func graceFullShutdown() {
	osKillSignalChan := make(chan os.Signal, 1)
	signal.Notify(osKillSignalChan, os.Interrupt, syscall.SIGTERM)
	killSignal := <-osKillSignalChan
	switch killSignal {
	case os.Interrupt:
		log.Info("got SIGINT...")
	case syscall.SIGTERM:
		log.Info("got SIGTERM...")
	}
}

func cronHandle(date time.Time, ec2Service *aws.EC2Service, instanceId string, isStart bool) bool {
	location, _ := time.LoadLocation("Europe/Moscow")
	date = time.Date(date.Year(), date.Month(), date.Day(), date.Hour(), date.Minute(), date.Second(), date.Nanosecond(), location)
	err := gocron.Every(1).Hour().From(&date).Do(doTask(ec2Service, instanceId, isStart))
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}

func doTask(ec2Service *aws.EC2Service, instanceId string, isStart bool) func() {
	return func() {
		log.Println(time.Now())
		log.Println("I am running task.")
		if isStart {
			err := ec2Service.StartInstance(&instanceId)
			if err != nil {
				log.Println(err)
			}
		} else {
			err := ec2Service.StopInstance(&instanceId)
			if err != nil {
				log.Println(err)
			}
		}
	}
}
