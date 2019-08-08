package main

import (
	"flag"

	"github.com/golang/glog"
	"github.com/inhzus/go-redis-impl/internal/pkg/client"
	"github.com/inhzus/go-redis-impl/internal/pkg/command"
	"github.com/inhzus/go-redis-impl/internal/pkg/token"
	"fmt"
)

func init() {
	flag.Parse()
	_ = flag.Set("stderrthreshold", "INFO")
}

func main() {
	cli := client.NewClient(&client.Option{})
	err := cli.Connect()
	if err != nil {
		glog.Error(err)
		return
	}
	_, _ = cli.Request(token.NewString(command.StringPing))
	_, _ = cli.Request(token.NewArray(
		token.NewString("set"), token.NewString("a"), token.NewBulked([]byte("what"))))
	_, _ = cli.Request(token.NewArray(
		token.NewString("get"), token.NewString("a")))
	_ = cli.Close()
	x := fmt.Sprintf("%v%v%v", 1, "%v", 3)
	fmt.Printf(x)
	//fmt.Printf(fmt.Sprintf("%v%v%v", 1, "%v", 3), 2)
}
