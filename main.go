package main

import (
	"flag"
	"github.com/chuccp/http2smtp/api"
	"github.com/chuccp/http2smtp/core"
	"github.com/chuccp/http2smtp/manage"
	"github.com/chuccp/http2smtp/schedule"
)

func main() {

	var webPort int
	var apiPort int
	flag.IntVar(&webPort, "web_port", 0, "web port")
	flag.IntVar(&apiPort, "api_port", 0, "api port")
	flag.Parse()
	smtp2Http := core.Create()
	smtp2Http.AddServer(manage.NewServer())
	smtp2Http.AddServer(api.NewServer())
	smtp2Http.AddServer(schedule.NewServer())
	smtp2Http.Start(webPort, apiPort)
}
