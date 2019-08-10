package main

import (
	"flag"
	"time"

	"github.com/golang/glog"
	"github.com/inhzus/go-redis-impl/internal/pkg/client"
	"github.com/inhzus/go-redis-impl/internal/pkg/command"
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

	cli.Submit(token.NewArray(
		token.NewString(command.CmdIncr), token.NewString("c")))
	cli.Submit(token.NewArray(
		token.NewString(command.CmdIncr), token.NewString("c")))
	_, _ = cli.Get("c")

	glog.Infof("time: %v", time.Now().Sub(startTime))
}
