package config

import (
	"flag"
	"fmt"
	"time"

	"github.com/inhzus/go-redis-impl/internal/pkg/client"
)

func Init() *client.Option {
	option := &client.Option{}
	var host, port string
	var readTimeout, writeTimeout int64
	flag.StringVar(&option.Proto, "protocol", "tcp", "protocol")
	flag.StringVar(&host, "host", "127.0.0.1", "host")
	flag.StringVar(&port, "port", "6389", "port")
	flag.Int64Var(&readTimeout, "rt", 0, "read timeout (ms), \"0\" represents unlimited")
	flag.Int64Var(&writeTimeout, "wt", 0, "write timeout (ms), \"0\" represents unlimited")
	flag.Parse()
	option.Addr = fmt.Sprintf("%s:%s", host, port)
	option.ReadTimeout = time.Duration(readTimeout) * time.Millisecond
	option.WriteTimeout = time.Duration(writeTimeout) * time.Millisecond
	return option
}
