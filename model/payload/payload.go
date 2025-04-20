package payload

import (
	"github.com/a-aslani/wotop.git"
)

type Payload struct {
	Data      any                   `json:"data"`
	Publisher wotop.ApplicationData `json:"publisher"`
	TraceID   string                `json:"traceId"`
}
