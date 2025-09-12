import { exec } from "child_process";
import fs from "fs";
import path from "path";

export function runDocker(problemConfig, uploadedFilePath, baseDir, callback) {
  try {
    // Prepare work directory
    const inputData = problemConfig.input;
    const workDir = path.join(baseDir, "../docker-scripts", `${Date.now()}`);
    fs.mkdirSync(workDir, { recursive: true });

    const codeFilePath = path.join(workDir, "solution.py");
    fs.renameSync(uploadedFilePath, codeFilePath);

    const inputFilePath = path.join(workDir, "input.txt");
    fs.writeFileSync(inputFilePath, inputData);

    // Secure Docker command with resource limits
    const dockerCommand = `
docker run --rm \
  --cpus="0.5" \
  --memory="128m" \
  --pids-limit=64 \
  --network=none \
  --read-only \
  -v ${workDir}:/app \
  ${problemConfig.image} \
  sh -c "${problemConfig.command}"
    `;

    console.log("Executing Docker command:", dockerCommand);

    exec(dockerCommand, { timeout: 5000 }, (error, stdout, stderr) => {
      // Cleanup work dir
      fs.rmSync(workDir, { recursive: true, force: true });

      if (error) {
        return callback({
          error: "Docker execution failed",
          details: error.message,
          stderr: stderr,
        });
      }

      callback(null, {
        actual_output: stdout.trim(),
        expected_output: problemConfig.expectedOutput,
        pass: stdout.trim() === problemConfig.expectedOutput,
      });
    });
  } catch (err) {
    callback({
      error: "Internal Docker execution error",
      details: err.message,
    });
  }
}
