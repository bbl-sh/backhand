import { exec } from "child_process";
import fs from "fs/promises";
import path from "path";
import os from "os";

// This function is now async and returns a Promise
export async function runDocker(problemConfig, uploadedFilePath) {
  let workDir; // Define here to access it in the 'finally' block

  try {
    // 1. Create a unique, temporary directory for this execution
    workDir = await fs.mkdtemp(path.join(os.tmpdir(), "code-run-"));

    // 2. Move the uploaded file into the temporary directory
    const codeFilePath = path.join(workDir, problemConfig.filename);
    await fs.rename(uploadedFilePath, codeFilePath);

    // 3. Create the input file
    const inputFilePath = path.join(workDir, "input.txt");
    await fs.writeFile(inputFilePath, problemConfig.input || "");

    // 4. Construct the Docker command
    const dockerCommand = `docker run --rm \
      --cpus="0.5" \
      --memory="128m" \
      --pids-limit=64 \
      --network=none \
      --read-only \
      -v "${workDir}":/app \
      ${problemConfig.image} \
      sh -c "${problemConfig.command}"`;

    console.log(`Executing in: ${workDir}`);

    // 5. Execute the command within a Promise
    return await new Promise((resolve, reject) => {
      exec(dockerCommand, { timeout: 10000 }, (error, stdout, stderr) => {
        if (error) {
          // If Docker times out or has a non-zero exit code
          console.error("Docker execution error:", stderr || error.message);
          return reject({
            message: "Execution Error",
            stderr: stderr || "Process timed out or crashed.",
            stdout: stdout,
          });
        }
        // Success
        const actualOutput = stdout.trim();
        resolve({
          pass: actualOutput === problemConfig.expectedOutput,
          actual_output: actualOutput,
          expected_output: problemConfig.expectedOutput,
        });
      });
    });
  } finally {
    // 6. ALWAYS clean up the temporary directory
    if (workDir) {
      await fs.rm(workDir, { recursive: true, force: true });
      console.log(`Cleaned up directory: ${workDir}`);
    }
  }
}
