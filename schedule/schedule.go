package schedule

import (
	"github.com/chuccp/http2smtp/core"
	"github.com/chuccp/http2smtp/db"
	"github.com/chuccp/http2smtp/util"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"log"
	"sync"
	"time"
)

type cronManage struct {
	cronMap map[uint]*Info
	cron    *cron.Cron
	lock    *sync.Mutex
	context *core.Context
	isStart bool
	isStop  bool
}
type Info struct {
	entryID    cron.EntryID
	Id         uint
	UpdateTime time.Time
	Key        string
}

func NewScheduleInfo(entryID cron.EntryID, schedule *db.Schedule) *Info {
	return &Info{
		entryID:    entryID,
		Id:         schedule.GetId(),
		Key:        schedule.Key(),
		UpdateTime: schedule.UpdateTime,
	}
}

func newCronManage(context *core.Context) *cronManage {
	return &cronManage{
		cronMap: make(map[uint]*Info),
		cron:    cron.New(cron.WithSeconds()),
		lock:    new(sync.Mutex),
		context: context,
		isStart: false,
		isStop:  false,
	}
}
func (cronM *cronManage) Start() {
	if cronM.context.GetDb() == nil {
		return
	}
	cronM.isStart = true
	cronM.cron.Start()
	go cronM.run()
}
func (cronM *cronManage) run() {
	for {
		if cronM.isStop {
			break
		}
		schedules, err := cronM.context.GetDb().GetScheduleModel().FindAllByUse()
		if err != nil {
			cronM.context.GetLog().Error("GetScheduleModel error", zap.Error(err))
			continue
		}
		addSchedule := make([]*db.Schedule, 0)
		updateSchedule := make([]*db.Schedule, 0)
		deleteSchedule := make([]*Info, 0)
		ids := make([]uint, len(schedules))
		for index, schedule := range schedules {
			ids[index] = schedule.GetId()
			info, ok := cronM.cronMap[schedule.GetId()]
			if ok {
				if schedule.UpdateTime.After(info.UpdateTime) || schedule.Key() != info.Key {
					updateSchedule = append(updateSchedule, schedule)
				}
			} else {
				addSchedule = append(addSchedule, schedule)
			}
		}
		for id, info := range cronM.cronMap {
			if !util.ContainsNumberAny(id, ids...) {
				deleteSchedule = append(deleteSchedule, info)
			}
		}
		log.Printf("schedule update %d add %d delete %d", len(updateSchedule), len(addSchedule), len(deleteSchedule))
		for _, schedule := range addSchedule {
			cronM.addSchedule(schedule)
		}
		for _, schedule := range updateSchedule {
			cronM.deleteSchedule(schedule.GetId())
			cronM.addSchedule(schedule)
		}
		for _, info := range deleteSchedule {
			cronM.deleteSchedule(info.Id)
		}
		time.Sleep(time.Second * 10)
	}
}
func (cronM *cronManage) deleteSchedule(id uint) {
	info, ok := cronM.cronMap[id]
	if ok {
		delete(cronM.cronMap, id)
		cronM.cron.Remove(info.entryID)
	}
}

func (cronM *cronManage) addSchedule(schedule *db.Schedule) {
	entryID, err := cronM.cron.AddFunc(schedule.Cron, func() {
		cronM.context.GetTokenService().SendApiCallMail(schedule)
	})
	if err != nil {
		cronM.context.GetLog().Error("cron start error", zap.String("cron", schedule.Cron), zap.Error(err))
	} else {
		cronM.context.GetLog().Info("cron start add", zap.String("cron", schedule.Cron))
		cronM.cronMap[schedule.GetId()] = NewScheduleInfo(entryID, schedule)
	}
}

func (cronM *cronManage) stop() {
	if !cronM.isStart {
		return
	}
	cronM.isStop = true
	cronM.cron.Stop()
}

type Server struct {
	cronManage *cronManage
	lock       *sync.Mutex
	request    *util.Request
	context    *core.Context
}

func NewServer() *Server {
	return &Server{
		lock:    new(sync.Mutex),
		request: util.NewRequest(),
	}
}

func (server *Server) Init(context *core.Context) {
	server.context = context
}
func (server *Server) Name() string {
	return "schedule"
}
func (server *Server) Start() {
	isInit := server.context.GetConfig().GetBooleanOrDefault("core", "init", false)
	isStart := server.context.GetConfig().GetBooleanOrDefault("schedule", "start", true)
	server.context.GetLog().Info("schedule_init", zap.Bool("isInit", isInit), zap.Bool("isStart", isStart))
	if isInit && isStart {
		server.lock.Lock()
		defer server.lock.Unlock()
		if server.cronManage != nil {
			server.cronManage.stop()
		}
		server.cronManage = newCronManage(server.context)
		server.cronManage.Start()
		server.context.GetLog().Info("start Schedule")
	}
}
