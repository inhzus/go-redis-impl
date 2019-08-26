package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/inhzus/go-redis-impl/internal/pkg/cds"
	"github.com/inhzus/go-redis-impl/internal/pkg/client"
	"github.com/inhzus/go-redis-impl/internal/pkg/token"
)

func getOption() *client.Option {
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

func splitIntoArray(line string) []string {
	var r = regexp.MustCompile(`[^\W"]+|"[^"]+"`)
	fi := r.FindAllString(line, -1)
	for i, s := range fi {
		if s[0] == '"' {
			fi[i] = s[1 : len(s)-1]
		}
	}
	return fi
}

func formStr(s string) string {
	return fmt.Sprintf("+%s\r\n", s)
}

func formBulked(s string) string {
	return fmt.Sprintf("$%d\r\n%s\r\n", len(s), s)
}

func formNum(s string) string {
	return fmt.Sprintf(":%s\r\n", s)
}

func formArray(a []string) []byte {
	data := []byte(fmt.Sprintf("*%d\r\n", len(a)))
	data = append(data, strings.Join(a, "")...)
	return data
}

func main() {
	opt := getOption()
	cli := client.NewClient(opt)
	err := cli.Connect()

	teardown := func() { fmt.Println(); cli.Close(); os.Exit(0) }
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sig
		teardown()
	}()

	if err != nil {
		fmt.Print(err.Error())
		return
	}
	var line string
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("%s[%d]>", opt.Addr, opt.Database)
		line, err = reader.ReadString('\n')
		if err != nil || len(line) <= 1 {
			teardown()
		}
		line = line[:len(line)-1]
		cmd := splitIntoArray(line)
		if len(cmd) < 1 {
			teardown()
		}
		var data []byte
		switch cmd[0] {
		case cds.Discard, cds.Exec, cds.Multi, cds.Ping, cds.Unwatch:
			cmd = cmd[:1]
		case cds.Desc, cds.Get, cds.Incr, cds.Watch:
			cmd = cmd[:2]
			cmd[1] = formStr(cmd[1])
		case cds.Select:
			cmd = cmd[:2]
			n, err := strconv.ParseInt(cmd[1], 10, 64)
			if err != nil {
				fmt.Printf("database index unable to parse")
				continue
			}
			opt.Database = int(n)
			cmd[1] = formNum(cmd[1])
		case cds.Set:
			cmd[1] = formStr(cmd[1])
			cmd[2] = formBulked(cmd[2])
			cnt := 3
			for i := cnt; i < len(cmd); i += 2 {
				switch cmd[i] {
				case cds.TimeoutSec, cds.TimeoutMilSec, cds.ExpireAtNano:
					if i+1 >= len(cmd) {
						fmt.Printf("missing argument of \"%s\"\n", cmd[i])
						continue
					}
					cmd[i] = formStr(cmd[i])
					cmd[i+1] = formNum(cmd[i+1])
				default:
					fmt.Printf("unrecognized argument: \"%s\"\n", cmd[i])
					continue
				}
			}
		default:
			fmt.Printf("unrecognized cmd: \"%s\"\n", cmd[0])
			continue
		}
		cmd[0] = formStr(cmd[0])
		//fmt.Printf("origin: %#v\n", line)
		data = formArray(cmd)
		//data = []byte(strings.Join(strings.Split(line, " "), "\r\n"))
		//fmt.Printf("cmd: %#v\n", string(data))
		_, err := cli.Conn.Write(data)
		//fmt.Print("write\n")
		if err != nil {
			fmt.Print(err.Error())
			continue
		}
		rsp, _ := token.Deserialize(cli.Conn)
		//fmt.Print("des\n")
		for _, t := range rsp {
			fmt.Println(t.Format())
		}
	}
}
