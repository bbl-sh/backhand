// handlers/test.go
package handlers

import (
	"strings"

	"github.com/john221wick/golang-backend/learn/internal/config"
	"github.com/john221wick/golang-backend/learn/internal/docker"
)

type TestInput struct {
	ProblemID int
	Code      []byte
}

type TestOutput struct {
	Success        bool
	Status         string // "Accepted", "Wrong Answer"
	ActualOutput   string
	ExpectedOutput string
	Error          string
}

// RunTest runs the code and checks if output matches expected
func RunTest(input TestInput) TestOutput {
	problem, exists := config.ProblemConfigs[input.ProblemID]
	if !exists {
		return TestOutput{
			Success: false,
			Status:  "Invalid Problem",
			Error:   "Problem not found",
		}
	}

	// Run in Docker
	actual, err := docker.RunInContainer(problem.Image, problem.Command, input.Code, []byte(problem.Input))
	if err != nil {
		return TestOutput{
			Success: false,
			Status:  "Runtime Error",
			Error:   err.Error(),
		}
	}

	// Normalize whitespace (optional: trim & normalize newlines)
	cleanActual := strings.TrimSpace(actual)
	cleanExpected := strings.TrimSpace(problem.ExpectedOutput)

	status := "Wrong Answer"
	if cleanActual == cleanExpected {
		status = "Accepted"
	}

	return TestOutput{
		Success:        cleanActual == cleanExpected,
		Status:         status,
		ActualOutput:   actual, // raw output
		ExpectedOutput: problem.ExpectedOutput,
	}
}
