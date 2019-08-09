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
	pipeline := cli.Pipeline()
	_, _ = pipeline.Ping()
	_, _ = pipeline.Ping()
	_, _ = pipeline.Exec()
	_, _ = pipeline.Ping()
	_, _ = pipeline.Get("a")
	_, _ = pipeline.Exec()
	cli.Close()

	glog.Infof("time: %v", time.Now().Sub(startTime))
}
