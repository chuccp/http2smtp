package main

import (
	"flag"
	"github.com/chuccp/http2smtp/api"
	"github.com/chuccp/http2smtp/core"
	"github.com/chuccp/http2smtp/manage"
	"github.com/chuccp/http2smtp/schedule"
	"github.com/chuccp/http2smtp/util"
)

func main() {

	var webPort int
	var apiPort int
	var storageRoot string
	flag.IntVar(&webPort, "web_port", 0, "web port")
	flag.IntVar(&apiPort, "api_port", 0, "api port")
	flag.StringVar(&storageRoot, "storage_root", "./", "storage root")
	flag.Parse()
	webPort = util.GetEnvIntOrDefault("web_port", webPort)
	apiPort = util.GetEnvIntOrDefault("api_port", apiPort)
	storageRoot = util.GetEnvOrDefault("storage_root", storageRoot)
	smtp2Http := core.Create()
	smtp2Http.AddServer(manage.NewServer())
	smtp2Http.AddServer(api.NewServer())
	smtp2Http.AddServer(schedule.NewServer())
	smtp2Http.Start(webPort, apiPort, storageRoot)
}
