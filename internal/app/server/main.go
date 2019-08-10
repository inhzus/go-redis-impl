package main

import (
	"flag"

	"github.com/go-redis/redis"
	"github.com/inhzus/go-redis-impl/internal/pkg/server"
)

func init() {
	flag.Parse()
	_ = flag.Set("stderrthreshold", "INFO")
}

func main() {
	s := server.NewServer(&server.Option{})
	_ = s.Serve()
	cli := redis.NewClient(&redis.Options{})
	cli.Set("x", 2, 0)
	//time.Sleep(time.Second)
	//s.Close()
}
