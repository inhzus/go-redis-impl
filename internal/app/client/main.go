package main

import (
	"flag"

	"github.com/inhzus/go-redis-impl/internal/pkg/client"
	"github.com/inhzus/go-redis-impl/internal/pkg/command"
	"github.com/inhzus/go-redis-impl/internal/pkg/token"
)

func init() {
	flag.Parse()
	_ = flag.Set("stderrthreshold", "INFO")
}

func main() {
	cli := client.NewClient(&client.Option{})
	_ = cli.Connect()

	cli.Request(token.NewString(command.StringPing))
	_ = cli.Close()
}
