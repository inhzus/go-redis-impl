package main

import (
	"flag"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/golang/glog"
	"github.com/inhzus/go-redis-impl/internal/pkg/server"
)

func parseDuration(str string) time.Duration {
	durationRegex := regexp.MustCompile(`(?P<years>\d+Y)?(?P<months>\d+M)?(?P<days>\d+D)?(?P<hours>\d+h)?(?P<minutes>\d+m)?(?P<seconds>\d+s)?`)
	matches := durationRegex.FindStringSubmatch(str)

	years := parseInt64(matches[1])
	months := parseInt64(matches[2])
	days := parseInt64(matches[3])
	hours := parseInt64(matches[4])
	minutes := parseInt64(matches[5])
	seconds := parseInt64(matches[6])

	hour := int64(time.Hour)
	minute := int64(time.Minute)
	second := int64(time.Second)
	return time.Duration(years*24*365*hour + months*30*24*hour + days*24*hour + hours*hour + minutes*minute + seconds*second)
}

func parseInt64(value string) int64 {
	if len(value) == 0 {
		return 0
	}
	parsed, err := strconv.Atoi(value[:len(value)-1])
	if err != nil {
		return 0
	}
	return int64(parsed)
}

func getOption() *server.Option {
	_ = flag.Set("stderrthreshold", "INFO")
	opt := &server.Option{}
	var host, port string
	var flushInterval, rewriteInterval string
	flag.StringVar(&opt.Proto, "protocol", "tcp", "protocol")
	flag.StringVar(&host, "host", "127.0.0.1", "host")
	flag.StringVar(&port, "port", "6389", "port")
	flag.BoolVar(&opt.Persist.Enable, "dr", false, "disable auto restore from persistence file")
	flag.BoolVar(&opt.Persist.SaveCopy, "es", false, "enable save persistence file with timestamp")
	flag.StringVar(&flushInterval, "fi", "1s", "flushing to aof interval, format: 1Y2M3D4h5m6s")
	flag.StringVar(&rewriteInterval, "ri", "1h", "rewriting rcl interval, format: 1Y2M3D4h5m6s")
	flag.Parse()
	opt.Addr = fmt.Sprintf("%s:%s", host, port)
	opt.Persist.FlushInr = parseDuration(flushInterval)
	opt.Persist.RewriteInr = parseDuration(rewriteInterval)
	opt.Persist.Enable = !opt.Persist.Enable
	return opt
}

func main() {
	opt := getOption()
	s := server.NewServer(opt)
	glog.Info(opt)
	s.Serve()
}
