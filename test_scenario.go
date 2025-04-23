package wotop

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestScenario represents a test case scenario for testing an Inport.
//
// Type Parameters:
//   - REQUEST: The type of the request object.
//   - RESPONSE: The type of the response object.
//   - OUTPORT: The type of the outport dependency.
type TestScenario[REQUEST, RESPONSE, OUTPORT any] struct {
	Name           string    // The name of the test case.
	InportRequest  REQUEST   // The input request to be passed to the Inport.
	InportResponse *RESPONSE // The expected response from the Inport.
	Outport        OUTPORT   // The outport dependency to be used in the test case.
	ExpectedError  error     // The expected error, if any, from the Inport execution.
}

// RunTestcaseScenarios runs a list of test scenarios for an Inport.
//
// This function executes each test scenario in parallel, invoking the provided
// Inport function with the given outport and request. It then asserts the
// response and error against the expected values.
//
// Type Parameters:
//   - REQUEST: The type of the request object.
//   - RESPONSE: The type of the response object.
//   - OUTPORT: The type of the outport dependency.
//
// Parameters:
//   - t: The testing object used to manage test execution.
//   - f: A function that takes an OUTPORT and returns an Inport instance.
//   - scenarioList: A variadic list of TestScenario objects to be executed.
func RunTestcaseScenarios[REQUEST, RESPONSE, OUTPORT any](t *testing.T, f func(o OUTPORT) Inport[REQUEST, RESPONSE], scenarioList ...TestScenario[REQUEST, RESPONSE, OUTPORT]) {

	t.Parallel() // Run the test cases in parallel.

	for _, tt := range scenarioList {

		t.Run(tt.Name, func(t *testing.T) {

			// Execute the Inport with the provided request and outport.
			res, err := f(tt.Outport).Execute(context.Background(), tt.InportRequest)

			// Assert the error if one is expected.
			if err != nil {
				assert.Equal(t, tt.ExpectedError, err, "Testcase name %s", tt.Name)
				return
			}

			// Assert the response if no error occurred.
			assert.Equal(t, tt.InportResponse, res, "Testcase name %s", tt.Name)

		})

	}

}
