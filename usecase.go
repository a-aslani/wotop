package wotop

import (
	"context"
	"fmt"
	"os"
)

// Inport defines a generic interface for use cases with a request and response type.
//
// Type Parameters:
//   - REQUEST: The type of the request object.
//   - RESPONSE: The type of the response object.
type Inport[REQUEST, RESPONSE any] interface {
	// Execute processes the given request and returns a response or an error.
	//
	// Parameters:
	//   - ctx: The context for managing request-scoped values, deadlines, and cancellations.
	//   - req: The request object of type REQUEST.
	//
	// Returns:
	//   - A pointer to the response object of type RESPONSE, or an error if the execution fails.
	Execute(ctx context.Context, req REQUEST) (*RESPONSE, error)
}

// GetInport retrieves and validates an Inport instance from a use case.
//
// This function ensures that the provided use case can be cast to the Inport interface
// with the specified request and response types. If the use case is invalid or cannot
// be cast, the function logs an error message and terminates the program.
//
// Type Parameters:
//   - Req: The type of the request object.
//   - Res: The type of the response object.
//
// Parameters:
//   - usecase: The use case to be cast to the Inport interface.
//   - err: An error object that, if non-nil, will cause the program to terminate.
//
// Returns:
//   - An Inport instance with the specified request and response types.
func GetInport[Req, Res any](usecase any, err error) Inport[Req, Res] {

	// Check if an error was provided and terminate the program if so.
	if err != nil {
		fmt.Printf("\n\n%s...\n\n", err.Error())
		os.Exit(0)
	}

	// Attempt to cast the use case to the Inport interface.
	inport, ok := usecase.(Inport[Req, Res])
	if !ok {
		// Log an error message and terminate the program if the cast fails.
		fmt.Printf("unable to cast to Inport\n")
		os.Exit(0)
	}
	return inport
}
