import { exec } from "child_process";

export function runDocker(workDir, problemConfig) {
  return new Promise((resolve, reject) => {
    const command = `
docker run --rm -v ${workDir}:/app --memory="100m" --cpus="0.5" ${problemConfig.image} sh -c "${problemConfig.command}"
    `;

    exec(command, { timeout: 10000 }, (error, stdout, stderr) => {
      if (error) {
        console.error("Docker error:", error.message);
        console.error("Stderr:", stderr);
        return reject(error);
      }
      return resolve({ stdout, stderr });
    });
  });
}
