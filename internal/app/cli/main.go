package main

import (
	"fmt"
	"time"

	"github.com/golang/glog"
	"github.com/inhzus/go-redis-impl/internal/app/cli/config"
	"github.com/inhzus/go-redis-impl/internal/pkg/client"
)

func main() {
	conf := config.Init()
	cli := client.NewClient(conf)
	_ = cli.Connect()
	//cli.Close()
	glog.Infof("config: %s", conf)
	pipe := cli.Pipeline()
	for i := 0; i < 20; i++ {
		pipe.Set(fmt.Sprintf("a%d", i), i, time.Hour)
	}
	_, _ = pipe.Exec()
	cli.Set("a", 1, 0)
	cli.Get("s")
	_ = cli
}
