package config

// ProblemConfig defines a coding problem
type ProblemConfig struct {
	Image          string // Docker image
	Command        string // Command to run
	Input          string // Input to feed via stdin/file
	ExpectedOutput string // Expected output (trimmed, newline-sensitive)
}

// ProblemConfigs maps problem ID to config
var ProblemConfigs = map[int]ProblemConfig{
	1: {
		Image:          "python-basic",
		Command:        "python /app/solution.py < /app/input.txt",
		Input:          "alice",
		ExpectedOutput: "Hello, alice!",
	},
	2: {
		Image:          "python-basic",
		Command:        "python /app/solution.py < /app/input.txt",
		Input:          "2\n3",
		ExpectedOutput: "5",
	},
}
