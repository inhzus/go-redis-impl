package server

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/golang/glog"
	"github.com/inhzus/go-redis-impl/internal/pkg/cds"
	"github.com/inhzus/go-redis-impl/internal/pkg/model"
	"github.com/inhzus/go-redis-impl/internal/pkg/proc"
	"github.com/inhzus/go-redis-impl/internal/pkg/token"
)

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func (s *Server) cloneData() (err error) {
	var file, dst *os.File
	file, err = os.OpenFile(s.option.Persist.CloneName, os.O_CREATE|os.O_RDWR, 0644)
	defer func() { _ = file.Close() }()
	if err != nil {
		return
	}
	// store initialization data
	var buffer bytes.Buffer
	var notEmpty bool
	for i := 0; i < s.option.DBCount; i++ {
		// notify data to freeze
		c := make(chan error)
		s.queue <- &model.ModTask{Cmd: proc.ModFreeze, DataIdx: i, Rsp: c}
		err = <-c
		if err != nil {
			continue
		}
		// select database
		t, _ := token.NewArray(token.NewString(cds.Select), token.NewInteger(int64(i))).Serialize()
		buffer.Write(t)
		// send channel to receive data
		dataCh := make(chan []byte)
		go s.proc.GenBin(i, dataCh)
		count := 0
		for d := range dataCh {
			count++
			buffer.Write(d)
			if buffer.Len() >= 2<<20 {
				_, err = buffer.WriteTo(file)
				if err != nil {
					return
				}
			}
		}
		// cancel lock, notify to move back to origin data
		s.queue <- &model.ModTask{Cmd: proc.ModMove, DataIdx: i, Rsp: c}
		_ = <-c
		if count == 0 {
			buffer.Reset()
		} else {
			notEmpty = true
			_, err = buffer.WriteTo(file)
			if err != nil {
				return
			}
		}
	}
	if !notEmpty {
		return
	}
	err = file.Sync()
	if err != nil {
		return
	}
	// empty aof after clone
	_, err = os.OpenFile(s.option.Persist.AppendName, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	if !s.option.Persist.SaveCopy {
		return
	}
	// copy the rcl file with timestamp
	dst, err = os.OpenFile(
		fmt.Sprintf("%v.%v", s.option.Persist.CloneName, time.Now().Format("0102-15:04:05")),
		os.O_CREATE|os.O_WRONLY,
		0644)
	if err != nil {
		return
	}
	defer func() { _ = dst.Close() }()
	_, err = file.Seek(0, 0)
	if err != nil {
		return
	}
	_, err = io.Copy(dst, file)
	if err != nil {
		return
	}
	err = dst.Sync()
	return err
}

// restore data from the aof and rcl
func (s *Server) restoreData() {
	rcl, err := os.OpenFile(s.option.Persist.CloneName, os.O_CREATE|os.O_RDONLY, 0644)
	checkErr(err)
	defer func() { _ = rcl.Close() }()
	reader := bufio.NewReader(rcl)
	cli := s.proc.NewClient(nil, 0, nil)
	// restore from the rcl
	ch := make(chan *token.Token)
	for t := range token.DeserializeGenerator(reader) {
		s.queue <- &model.CmdTask{Cli: cli, Req: t.T, Rsp: ch}
		<-ch
	}
	// then restore from the aof
	aof, err := os.OpenFile(s.option.Persist.AppendName, os.O_CREATE|os.O_RDWR, 0644)
	checkErr(err)
	defer func() { _ = rcl.Close() }()
	reader = bufio.NewReader(aof)
	cli = s.proc.NewClient(nil, 0, nil)
	for t := range token.DeserializeGenerator(reader) {
		s.queue <- &model.CmdTask{Cli: cli, Req: t.T, Rsp: ch}
		<-ch
	}
	//_ = aof.Truncate(0)
	//_ = aof.Sync()
}

func (s *Server) persistence() {
	// store the db to file instantly after server started
	checkErr(s.cloneData())
	// clone the whole data to the rcl periodically
	go func() {
		rewriteTicker := time.NewTicker(s.option.Persist.RewriteInr)
		for {
			<-rewriteTicker.C
			if err := s.cloneData(); err != nil {
				glog.Errorf("clone data: %v", err.Error())
			}
		}
	}()

	file, err := os.OpenFile(s.option.Persist.AppendName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	checkErr(err)
	var idx int
	flushTicker := time.NewTicker(s.option.Persist.FlushInr)
	var buffer bytes.Buffer
	for {
		select {
		// flush the aof file periodically
		case <-flushTicker.C:
			if buffer.Len() > 0 {
				_, err = buffer.WriteTo(file)
				if err != nil {
					glog.Errorf("aof write to file: %v", err.Error())
				}
				err = file.Sync()
				if err != nil {
					glog.Errorf("file sync: %v", err.Error())
				}
			}
			// receive the set msgs from the model clients and sync them to the aof
		case m := <-s.setCh:
			d, _ := m.T.Serialize()
			if m.Idx != idx {
				d, _ = token.NewArray(token.NewString(cds.Select), token.NewInteger(int64(m.Idx))).Serialize()
				buffer.Write(d)
			}
			buffer.Write(d)
		}
	}
}
