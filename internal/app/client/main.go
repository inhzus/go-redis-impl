package main

import (
	"flag"
	"time"

	"github.com/golang/glog"
	"github.com/inhzus/go-redis-impl/internal/pkg/client"
)

func init() {
	flag.Parse()
	_ = flag.Set("stderrthreshold", "INFO")
}

func main() {
	startTime := time.Now()
	cli := client.NewClient(&client.Option{})
	_ = cli.Connect()

	cli.Set("c", 2, time.Second)
	cli.Desc("c")
	cli.Get("c")
	<-time.After(time.Second)
	cli.Get("c")
	glog.Infof("time: %v", time.Now().Sub(startTime))
}
