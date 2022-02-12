package domain

import (
	"time"
)

type Task struct {
	Id    string
	Name  string
	Start time.Time
	Stop  time.Time
}

type TaskDto struct {
	Id    string `json:"instanceId"`
	Name  string `json:"instanceName"`
	Start string `json:"start"`
	Stop  string `json:"stop"`
}
