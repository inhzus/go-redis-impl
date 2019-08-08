package main

import (
	"flag"

	"github.com/inhzus/go-redis-impl/internal/pkg/server"
)

func init() {
	flag.Parse()
	_ = flag.Set("stderrthreshold", "INFO")
}

func main() {
	s := server.NewServer(&server.Option{})
	_ = s.Serve()
	//time.Sleep(time.Second)
	//s.Close()
}
