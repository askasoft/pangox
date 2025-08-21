package xschs

import (
	"fmt"

	"github.com/askasoft/pango/cog/linkedhashmap"
	"github.com/askasoft/pango/ini"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pango/sch"
)

var Schedules linkedhashmap.LinkedHashMap[string, func()]

func Register(name string, callback func()) {
	Schedules.Set(name, callback)
}

func InitScheduler() error {
	sch.Default().Logger = log.GetLogger("SCH")

	for it := Schedules.Iterator(); it.Next(); {
		name := it.Key()
		callback := it.Value()

		cron := ini.GetString("task", name)
		if cron == "" {
			sch.Schedule(name, sch.ZeroTrigger, callback)
		} else {
			ct, err := sch.NewCronTrigger(cron)
			if err != nil {
				return fmt.Errorf("invalid task '%s' cron: %w", name, err)
			}
			log.Infof("Schedule Task %s: %s", name, cron)
			sch.Schedule(name, ct, callback)
		}
	}

	return nil
}

func ReScheduler() {
	for _, name := range Schedules.Keys() {
		cron := ini.GetString("task", name)
		task, ok := sch.GetTask(name)
		if !ok {
			log.Errorf("Failed to find task %s", name)
			continue
		}

		if cron == "" {
			task.Stop()
		} else {
			redo := true
			if ct, ok := task.Trigger.(*sch.CronTrigger); ok {
				redo = (ct.Cron() != cron)
			}

			if redo {
				ct, err := sch.NewCronTrigger(cron)
				if err != nil {
					log.Errorf("Invalid task '%s' cron: %v", name, err)
				} else {
					log.Infof("Reschedule Task %s: %s", name, cron)
					task.Stop()
					task.Trigger = ct
					task.Start()
				}
			}
		}
	}
}
