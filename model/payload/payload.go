package payload

import (
	"github.com/a-aslani/wotop/wotop"
)

type Payload struct {
	Data      any         `json:"data"`
	Publisher wotop.wotop `json:"publisher"`
	TraceID   string      `json:"traceId"`
}
