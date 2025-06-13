package main

import (
	"flag"
	"github.com/chuccp/smtp2http/api"
	"github.com/chuccp/smtp2http/core"
	"github.com/chuccp/smtp2http/manage"
	"github.com/chuccp/smtp2http/unzip"
	"log"
	"strings"
)

func main() {

	var webPort int
	var apiPort int
	var unzipFile string
	flag.IntVar(&webPort, "web_port", 0, "web port")
	flag.IntVar(&apiPort, "api_port", 0, "api port")
	flag.StringVar(&unzipFile, "unzip", "", "unzip file dir")
	flag.Parse()
	if len(unzipFile) > 2 {
		vs := strings.Split(unzipFile, " ")
		err := unzip.Unzip(vs[0], vs[1])
		if err != nil {
			log.Panicln(err)
		}
	}
	smtp2Http := core.Create()
	smtp2Http.AddServer(manage.NewServer())
	smtp2Http.AddServer(api.NewServer())
	smtp2Http.Start(webPort, apiPort)
}
