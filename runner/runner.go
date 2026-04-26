package runner

import (
	"sync"

	"github.com/chuccp/go-web-frame/component/schedule"
	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/log"
	"github.com/chuccp/http2smtp/model"
	"github.com/chuccp/http2smtp/service"
	atomic2 "go.uber.org/atomic"
	"go.uber.org/zap"
)

type ScheduleRunner struct {
	scheduleService *service.ScheduleService
	tokenService    *service.TokenService
	schedule        *schedule.Schedule
	knownIds        map[uint]bool
	mu              sync.Mutex
	isRun           *atomic2.Bool
}

func (r *ScheduleRunner) Init(c *core.Context) error {
	r.scheduleService = core.GetService[*service.ScheduleService](c)
	r.tokenService = core.GetService[*service.TokenService](c)
	r.schedule = core.GetRunner[*schedule.Schedule](c)
	r.knownIds = make(map[uint]bool)
	r.isRun = atomic2.NewBool(false)
	return nil
}

func (r *ScheduleRunner) Run() error {
	r.loadAndSyncSchedules()
	_, err := r.schedule.AddFunc("0 */1 * * * ?", func(c *core.Context) {
		r.loadAndSyncSchedules()
	})
	if err != nil {
		return err
	}
	return nil
}

func (r *ScheduleRunner) loadAndSyncSchedules() {

	if !r.isRun.CompareAndSwap(false, true) {
		return
	}
	defer func() {
		r.isRun.Swap(false)
	}()
	r.mu.Lock()
	defer r.mu.Unlock()
	schedules, err := r.scheduleService.FindAll()
	if err != nil {
		log.Error("load schedules from database failed", zap.Error(err))
		return
	}
	activeIds := make(map[uint]bool)
	for _, sd := range schedules {
		activeIds[sd.Id] = true
		if !sd.IsUse {
			r.schedule.StopIdFunc(sd.Id)
			continue
		}
		key := sd.Key()
		sdCopy := sd
		_, _, err := r.schedule.AddIdOrReplaceKeyFunc(sd.Id, key, sd.Cron, func(c *core.Context) {
			r.executeSchedule(sdCopy)
		})
		if err != nil {
			log.Error("add schedule task failed", zap.Uint("scheduleId", sd.Id), zap.String("scheduleName", sd.Name), zap.Error(err))
		}
	}

	// Stop tasks that no longer exist in database
	for id := range r.knownIds {
		if !activeIds[id] {
			r.schedule.StopIdFunc(id)
		}
	}
	r.knownIds = activeIds
}

func (r *ScheduleRunner) executeSchedule(sd *model.Schedule) {
	err := r.tokenService.SendApiCallMail(sd)
	if err != nil {
		log.Error("execute schedule failed", zap.Uint("scheduleId", sd.Id), zap.String("scheduleName", sd.Name), zap.Error(err))
	}
}
