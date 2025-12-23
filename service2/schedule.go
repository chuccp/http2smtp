package service2

import "github.com/chuccp/go-web-frame/core"

type ScheduleService struct {
	context *core.Context
}

func (l *ScheduleService) Name() string {
	return "ScheduleService"
}
func (l *ScheduleService) Init(context *core.Context) {
	l.context = context
}
