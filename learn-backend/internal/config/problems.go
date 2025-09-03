// internal/config/problems.go
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
		Input:          "",
		ExpectedOutput: "5050",
	},

	2: {
		Image:          "python-basic",
		Command:        "python /app/solution.py < /app/input.txt",
		Input:          "1\n0\n3\na\n4\n0\nEND",
		ExpectedOutput: "3",
	},

	3: {
		Image:          "python-basic",
		Command:        "python /app/solution.py < /app/input.txt",
		Input:          "",
		ExpectedOutput: "720",
	},

	4: {
		Image:          "python-basic",
		Command:        "python /app/solution.py < /app/input.txt",
		Input:          "",
		ExpectedOutput: "5",
	},

	5: {
		Image:          "python-basic",
		Command:        "python /app/solution.py < /app/input.txt",
		Input:          "",
		ExpectedOutput: "8",
	},

	6: {
		Image:          "python-basic",
		Command:        "python /app/solution.py < /app/input.txt",
		Input:          "5",
		ExpectedOutput: "120",
	},

	7: {
		Image:          "python-basic",
		Command:        "python /app/solution.py < /app/input.txt",
		Input:          "Hello, World!\n",
		ExpectedOutput: "13",
	},

	8: {
		Image:          "python-basic",
		Command:        "python /app/solution.py < /app/input.txt",
		Input:          "3",
		ExpectedOutput: "6",
	},

	9: {
		Image:          "python-basic",
		Command:        "python /app/solution.py < /app/input.txt",
		Input:          "4",
		ExpectedOutput: "3",
	},

	10: {
		Image:          "python-basic",
		Command:        "python /app/solution.py < /app/input.txt",
		Input:          "apple",
		ExpectedOutput: "Hello, apple!",
	},
}
