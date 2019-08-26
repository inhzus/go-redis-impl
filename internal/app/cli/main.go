package main

import (
	"github.com/golang/glog"
	"github.com/inhzus/go-redis-impl/internal/app/cli/config"
	"github.com/inhzus/go-redis-impl/internal/pkg/client"
)

func main() {
	conf := config.Init()
	cli := client.NewClient(conf)
	glog.Infof("config: %s", conf)
	_ = cli.Connect()
	cli.Get("s")
	_ = cli
}
