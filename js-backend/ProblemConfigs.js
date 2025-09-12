// ProblemConfigs.js
export const ProblemConfigs = {
  1: {
    image: "python-basic-1",
    command: "python /app/solution.py < /app/input.txt",
    input: "",
    expectedOutput: "5050",
  },
  2: {
    image: "python-basic-1",
    command: "python /app/solution.py < /app/input.txt",
    input: "1\n0\n3\na\n4\n0\nEND",
    expectedOutput: "3",
  },
  // Add other problems here...
  10: {
    image: "python-basic-1",
    command: "python /app/solution.py < /app/input.txt",
    input: "apple",
    expectedOutput: "Hello, apple!",
  },
};
