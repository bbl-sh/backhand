import express from "express";
import multer from "multer";
import { exec } from "child_process";
import PocketBase from "pocketbase";
import { ProblemConfigs } from "./ProblemConfigs.js";
import path from "path";
import { fileURLToPath } from "url";
import fs from "fs";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const app = express();
const port = 3000;

// Initialize PocketBase client
const pb = new PocketBase("http://13.232.198.0:8890");

// Multer setup
const upload = multer({ dest: "uploads/" });

// Middleware
app.use(express.json());

/**
 * === Login Route ===
 * POST /login
 * Body: { email, password }
 */
app.post("/login", async (req, res) => {
  try {
    const { email, password } = req.body;

    if (!email || !password) {
      throw new Error("Email and password are required");
    }

    const authData = await pb
      .collection("users")
      .authWithPassword(email, password);

    return res.json({
      status: "Success",
      email: authData.record.email,
      id: authData.record.id,
      token: authData.token,
    });
  } catch (err) {
    console.error("Login error:", err.message);
    return res.status(400).json({
      status: "Error",
      message: err.message,
    });
  }
});

/**
 * === Challenge Route ===
 * POST /challenge01
 * Form-data: { challengeId, challengeName, problemId, code (file) }
 * Headers: { Authorization: Bearer <token>, X-User-Email: <email> }
 */
app.post("/challenge01", upload.single("code"), async (req, res) => {
  try {
    const { challengeId, challengeName, problemId } = req.body;

    console.log("Received challenge request:", req.body);

    const problemConfig = ProblemConfigs[problemId];
    if (!problemConfig) {
      throw new Error("Invalid problemId");
    }

    // Validate headers
    const authHeader = req.headers["authorization"];
    const userEmail = req.headers["x-user-email"];
    if (!authHeader || !userEmail) {
      throw new Error("Missing Authorization or X-User-Email header");
    }

    const token = authHeader.split(" ")[1];
    if (!token) {
      throw new Error("Invalid Authorization header format");
    }

    // ✅ Save token to pb.authStore and refresh
    pb.authStore.save(token, null);

    let userRecord;
    try {
      const authData = await pb.collection("users").authRefresh();
      userRecord = authData.record;

      if (!userRecord || userRecord.email !== userEmail) {
        throw new Error("User not found or email mismatch");
      }
    } catch (err) {
      throw new Error(
        "Authentication failed: " +
          (err.response?.data?.message || err.message),
      );
    }

    console.log("Authentication successful for:", userEmail);

    // Check uploaded file
    if (!req.file) {
      throw new Error("No file uploaded");
    }

    const filePath = req.file.path;
    console.log("Uploaded file path:", filePath);

    // Prepare work directory
    const inputData = problemConfig.input;
    const workDir = path.join(__dirname, "docker-scripts", `${Date.now()}`);
    fs.mkdirSync(workDir, { recursive: true });

    const codeFilePath = path.join(workDir, "solution.py");
    fs.renameSync(filePath, codeFilePath);

    const inputFilePath = path.join(workDir, "input.txt");
    fs.writeFileSync(inputFilePath, inputData);

    // Docker command
    const dockerCommand = `
docker run --rm -v ${workDir}:/app ${problemConfig.image} sh -c "${problemConfig.command}"
    `;

    console.log("Executing Docker command:", dockerCommand);

    exec(dockerCommand, (error, stdout, stderr) => {
      // Cleanup work dir
      fs.rmSync(workDir, { recursive: true, force: true });

      if (error) {
        console.error("Docker execution error:", error.message);
        console.error("Stderr:", stderr);
        return res.status(500).json({
          error: "Docker execution failed",
          details: error.message,
          stderr: stderr,
        });
      }

      console.log("Docker output:", stdout);

      return res.json({
        status: "Success",
        user_email: userEmail,
        actual_output: stdout.trim(),
        expected_output: problemConfig.expectedOutput,
        pass: stdout.trim() === problemConfig.expectedOutput,
        timestamp: new Date().toISOString(),
      });
    });
  } catch (err) {
    console.error("Error in challenge handler:", err.message);
    return res.status(400).json({
      status: "Error",
      message: err.message,
    });
  }
});

app.listen(port, () => {
  console.log(`Server running at http://localhost:${port}`);
});
