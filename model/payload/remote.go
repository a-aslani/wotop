package payload

import (
	"github.com/a-aslani/wotop/wotop"
)

type Args struct {
	Type      string      `json:"type"`
	Data      any         `json:"data"`
	Publisher wotop.wotop `json:"publisher"`
	TraceID   string      `json:"trace_id"`
}

type Reply struct {
	Success      bool
	ErrorMessage string
	Data         any
}
