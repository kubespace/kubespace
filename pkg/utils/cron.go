package utils

import (
	"github.com/robfig/cron/v3"
	"time"
)

func NextTriggerTime(spec string) (t time.Time, err error) {
	sch, err := cron.ParseStandard(spec)
	if err != nil {
		return
	}
	return sch.Next(time.Now()), nil
}
