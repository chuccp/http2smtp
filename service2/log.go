package service2

import "github.com/chuccp/go-web-frame/core"

type LogService struct {
	context *core.Context
}

func (l *LogService) Name() string {
	return "LogService"
}
func (l *LogService) Init(context *core.Context) {
	l.context = context
}
