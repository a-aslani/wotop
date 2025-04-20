package payload

import "github.com/a-aslani/wotop.git"

type Args struct {
	Type      string                `json:"type"`
	Data      any                   `json:"data"`
	Publisher wotop.ApplicationData `json:"publisher"`
	TraceID   string                `json:"trace_id"`
}

type Reply struct {
	Success      bool
	ErrorMessage string
	Data         any
}
