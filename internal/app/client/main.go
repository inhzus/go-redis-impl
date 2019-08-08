package main

import (
	"flag"
	"time"

	"github.com/golang/glog"
	"github.com/inhzus/go-redis-impl/internal/pkg/client"
	"github.com/inhzus/go-redis-impl/internal/pkg/token"
)

func init() {
	flag.Parse()
	_ = flag.Set("stderrthreshold", "INFO")
}

func main() {
	startTime := time.Now()
	cli := client.NewClient(&client.Option{})
	_ = cli.Connect()
	_ = cli.Request(
		token.NewArray(
			token.NewArray(token.NewString("ping")),
			token.NewArray(
				token.NewString("get"), token.NewString("a"))))
	cli.Close()

	glog.Infof("time: %v", time.Now().Sub(startTime))
}
