package transport

import (
	"aws/internal/common/domain"
	"aws/internal/server/application/aws"
	"aws/internal/server/domain/task"
	"encoding/json"
	"github.com/jasonlvhit/gocron"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type TaskController struct {
	repo       task.Repo
	ec2Service aws.EC2Service
}

func NewTaskController(repo *task.Repo, ec2Service *aws.EC2Service) *TaskController {
	t := new(TaskController)
	t.repo = *repo
	t.ec2Service = *ec2Service
	return t
}

func (c *TaskController) AddTask(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	if (*req).Method == "OPTIONS" {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	all, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), 400)
		return
	}
	taskDto := domain.TaskDto{}

	err = json.Unmarshal(all, &taskDto)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), 400)
		return
	}

	startDate, err := time.Parse(time.RFC3339, taskDto.Start+":00.000Z")
	stopDate, err := time.Parse(time.RFC3339, taskDto.Stop+":00.000Z")
	err = c.repo.CheckExist(taskDto.Id)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), 400)
		return
	}

	t := domain.Task{
		Id:    taskDto.Id,
		Name:  taskDto.Name,
		Start: startDate,
		Stop:  stopDate,
	}
	err = c.repo.Store(t)

	log.Println("add new task", taskDto, startDate)
	location, err := time.LoadLocation("Europe/Moscow")
	startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), startDate.Hour(), startDate.Minute(), startDate.Second(), startDate.Nanosecond(), location)
	err = gocron.Every(1).Hour().From(&startDate).Do(c.doTask(t.Id, true))
	if err != nil {
		log.Println(err)
		return
	}

	startDate = time.Date(stopDate.Year(), stopDate.Month(), stopDate.Day(), stopDate.Hour(), stopDate.Minute(), stopDate.Second(), stopDate.Nanosecond(), location)
	err = gocron.Every(1).Hour().From(&startDate).Do(c.doTask(t.Id, false))
	if err != nil {
		log.Println(err)
		return
	}
}

func (c *TaskController) doTask(instanceId string, isStart bool) func() {
	return func() {
		log.Println("I am running task.")
		if isStart {
			err := c.ec2Service.StartInstance(&instanceId)
			if err != nil {
				log.Println(err)
			}
		} else {
			err := c.ec2Service.StopInstance(&instanceId)
			if err != nil {
				log.Println(err)
			}
		}
	}
}
