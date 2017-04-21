package redis

import (
	"container/list"
	"errors"
	"fmt"
	"globaltedinc/framework/glog"
	"sync"
	"time"

	"github.com/mediocregopher/radix.v2/redis"
)

const (
	SimpleStr redis.RespType = 1 << iota
	BulkStr
	IOErr  // An error which prevented reading/writing, e.g. connection close
	AppErr // An error returned by redis, e.g. WRONGTYPE
	Int
	Array
	Nil

	// Str combines both SimpleStr and BulkStr, which are considered strings to
	// the Str() method.  This is what you want to give to IsType when
	// determining if a response is a string
	Str = SimpleStr | BulkStr

	// Err combines both IOErr and AppErr, which both indicate that the Err
	// field on their Resp is filled. To determine if a Resp is an error you'll
	// most often want to simply check if the Err field on it is nil
	Err = IOErr | AppErr
)

const maxReconnectInterval = 1 * time.Second

// Redis wrapper
type Redis struct {
	addr    string
	timeout uint32

	list  *list.List
	mutex sync.Mutex

	beginInvalidTime time.Time
}

// Open timeout: ms
func (r *Redis) Open(addr string, timeout uint32) {
	r.addr = addr
	r.timeout = timeout
	r.list = list.New()
}

func (r *Redis) Close() {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	for r.list.Len() > 0 {
		item := r.list.Back()
		c := item.Value.(*redis.Client)
		c.Close()
		c = nil
		r.list.Remove(item)
	}

	r.list = nil
}

func (r *Redis) createContext() *redis.Client {

	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.list.Len() > 0 {
		//glog.Info("get connection from list.")
		c := r.list.Back()
		r.list.Remove(c)
		return c.Value.(*redis.Client)
	}

	now := time.Now()
	if now.Before(r.beginInvalidTime.Add(maxReconnectInterval)) {
		return nil
	}

	client, err := redis.DialTimeout("tcp", r.addr, time.Duration(uint64(r.timeout))*time.Millisecond)
	if err != nil {
		glog.Info("Failed to call redis.DialTimeout. addr: ", r.addr, ", timeout: ", r.timeout, ", error: ", err)
		return nil
	}
	//glog.Error("create connection.")
	return client
}

func (r *Redis) releaseContext(c *redis.Client, active bool) {
	if !active {
		//glog.Error("close connection")
		c.Close()
		c = nil
		return
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()
	//glog.Info("put connection to list.")
	r.list.PushBack(c)
}

type Resp struct{ redis.Resp }

func (r *Redis) Cmd(cmd string, args ...interface{}) *Resp {
	c := r.createContext()
	if c == nil {
		return &Resp{redis.Resp{Err: errors.New(fmt.Sprintf("Cannot get a redis context. len: %d", r.list.Len()))}}
	}

	ret := &Resp{*c.Cmd(cmd, args)}
	r.releaseContext(c, ret.Err == nil)
	if ret.Err != nil {
		glog.Error("Failed to execute cmd. len: ", r.list.Len())
		glog.Error("args: ", args)
		glog.Error("error: ", ret.Err)
	}
	return ret
}

func (r *Redis) LoadScript(script string) (string, error) {
	ret := r.Cmd("SCRIPT", "LOAD", script)
	return ret.Str()
}
