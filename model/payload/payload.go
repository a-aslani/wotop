package payload

import (
	"github.com/a-aslani/wotop"
)

type Payload struct {
	Data      any                   `json:"data"`
	Publisher wotop.ApplicationData `json:"publisher"`
	TraceID   string                `json:"traceId"`
}
