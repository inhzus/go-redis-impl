package main

import (
	"flag"
	"strconv"
	"sync"
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
	wg := &sync.WaitGroup{}
	startTime := time.Now()
	for i := 0; i < 250; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cli := client.NewClient(&client.Option{})
			_ = cli.Connect()
			for j := 0; j < 100; j++ {
				rsp := cli.Request(token.NewArray(
					token.NewString("set"), token.NewString("a"),
					token.NewBulked([]byte(strconv.FormatInt(int64(i*j), 10)))))
				if rsp.Label == token.LabelError {
					glog.Errorf("error")
				}
				rsp = cli.Request(token.NewArray(
					token.NewString("get"), token.NewString("a")))
				if rsp.Label == token.LabelError {
					glog.Errorf("error")
				}
			}
			cli.Close()
		}()
	}
	wg.Wait()
	glog.Infof("time: %v", time.Now().Sub(startTime))
}
