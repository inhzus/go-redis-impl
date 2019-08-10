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
	cli := client.NewClient(&client.Option{
		ReadTimeout:  time.Millisecond / 2,
		WriteTimeout: time.Millisecond / 2})
	_ = cli.Connect()

	cli.Set("c", 2)
	cli.Desc("c")
	cli.Get("c")

	glog.Infof("time: %v", time.Now().Sub(startTime))
}
