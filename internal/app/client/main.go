package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/golang/glog"
	"github.com/inhzus/go-redis-impl/internal/pkg/client"
)

func init() {
	flag.Parse()
	_ = flag.Set("stderrthreshold", "INFO")
}

func main() {
	cli := client.NewClient(&client.Option{})
	_ = cli.Connect()

	startTime := time.Now()
	pipe := cli.Pipeline()
	for i := 0; i < 20; i++ {
		pipe.Set(fmt.Sprintf("a%d", i), i, time.Hour)
	}
	_, _ = pipe.Exec()
	cli.Set("a", 1, 0)
	glog.Infof("time: %v", time.Now().Sub(startTime))
}
