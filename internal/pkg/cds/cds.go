/*
The full name of the package is `command string`.
Extracting these constants from package proc is a must in my opinion
because they're either used by package client.
Client that imports processor of a server is weird in my mind.
*/
package cds

// command string
const (
	Desc    = "desc"
	Discard = "discard"
	Exec    = "exec"
	Get     = "get"
	Incr    = "incr"
	Multi   = "multi"
	Select  = "select"
	Set     = "set"
	Ping    = "ping"
	Unwatch = "unwatch"
	Watch   = "watch"
)

// argument string
const (
	TimeoutSec    = "EX"
	TimeoutMilSec = "PX"
	ExpireAtNano  = "PT"
)
