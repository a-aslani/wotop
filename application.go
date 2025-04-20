package wotop

import (
	"github.com/a-aslani/wotop/wotop_util"
	"time"
)

// Runner defines a generic interface for running a task with a given configuration.
//
// Type Parameters:
//   - T: The type of the configuration object.
type Runner[T any] interface {
	// Run executes the task with the provided configuration.
	//
	// Parameters:
	//   - cfg: A pointer to the configuration object of type T.
	//
	// Returns:
	//   - An error if the task execution fails, otherwise nil.
	Run(cfg *T) error
}

// ApplicationData represents metadata about the application instance.
//
// Fields:
//   - AppName: The name of the application.
//   - AppInstanceID: A unique identifier for the application instance.
//   - StartTime: The start time of the application instance in the format "YYYY-MM-DD HH:MM:SS".
type ApplicationData struct {
	AppName       string `json:"app_name"`        // The name of the application.
	AppInstanceID string `json:"app_instance_id"` // A unique identifier for the application instance.
	StartTime     string `json:"start_time"`      // The start time of the application instance.
}

// NewApplicationData creates a new ApplicationData instance with the given application name.
//
// Parameters:
//   - appName: The name of the application.
//
// Returns:
//   - An ApplicationData instance populated with the application name, a generated instance ID, and the current start time.
func NewApplicationData(appName string) ApplicationData {
	return ApplicationData{
		AppName:       appName,
		AppInstanceID: wotop_util.GenerateID(4),                 // Generate a unique 4-character ID for the application instance.
		StartTime:     time.Now().Format("2006-01-02 15:04:05"), // Set the current time as the start time.
	}
}
