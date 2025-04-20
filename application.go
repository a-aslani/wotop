package wotop

import (
	"github.com/a-aslani/wotop.git/wotop_util"
	"time"
)

type Runner[T any] interface {
	Run(cfg *T) error
}

type ApplicationData struct {
	AppName       string `json:"app_name"`
	AppInstanceID string `json:"app_instance_id"`
	StartTime     string `json:"start_time"`
}

func NewApplicationData(appName string) ApplicationData {
	return ApplicationData{
		AppName:       appName,
		AppInstanceID: wotop_util.GenerateID(4),
		StartTime:     time.Now().Format("2006-01-02 15:04:05"),
	}
}
