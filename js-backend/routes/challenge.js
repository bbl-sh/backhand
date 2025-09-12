import express from "express";
import multer from "multer";
import { exec } from "child_process";
import path from "path";
import { fileURLToPath } from "url";
import fs from "fs";
import jwt from "jsonwebtoken";
import pb from "../config/pocketbase.js"; // Re-exported PocketBase instance
import { ProblemConfigs } from "../ProblemConfigs.js";
import { decodeToken } from "../utils/jwt.js"; // 👈 Import your JWT utility

const router = express.Router();
const upload = multer({ dest: "uploads/" });

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

// POST /challenge/challenge01
router.post("/challenge01", upload.single("code"), async (req, res) => {
  try {
    const { challengeId, problemId } = req.body;
    const authHeader = req.headers["authorization"];
    const userEmail = req.headers["x-user-email"];

    if (!authHeader || !userEmail) {
      return res
        .status(400)
        .json({ error: "Missing Authorization or X-User-Email header" });
    }

    const token = authHeader.split(" ")[1];
    if (!token) {
      return res
        .status(400)
        .json({ error: "Invalid Authorization header format" });
    }

    const decoded = decodeToken(token); // 👈 Use your utility
    if (!decoded || !decoded.id) {
      return res.status(401).json({ error: "Invalid or malformed token" });
    }

    // Verify user exists and email matches
    const userRecord = await pb.collection("users").getOne(decoded.id);
    if (!userRecord || userRecord.email !== userEmail) {
      return res.status(403).json({ error: "User authentication failed" });
    }

    console.log("✅ Authenticated user:", userEmail);

    // Validate file upload
    if (!req.file) {
      return res.status(400).json({ error: "No file uploaded" });
    }

    const filePath = req.file.path;
    const problemConfig = ProblemConfigs[problemId];

    if (!problemConfig) {
      return res.status(400).json({ error: "Invalid problemId" });
    }

    const workDir = path.join(__dirname, "../docker-scripts", `${Date.now()}`);
    fs.mkdirSync(workDir, { recursive: true });

    const codeFilePath = path.join(workDir, "solution.py");
    fs.renameSync(filePath, codeFilePath);

    const inputFilePath = path.join(workDir, "input.txt");
    fs.writeFileSync(inputFilePath, problemConfig.input);

    const dockerCommand = `
docker run --rm -v ${workDir}:/app ${problemConfig.image} sh -c "${problemConfig.command}"
    `.trim();

    console.log("Executing Docker command:", dockerCommand);

    exec(dockerCommand, (error, stdout, stderr) => {
      // Cleanup
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

      const passed = stdout.trim() === problemConfig.expectedOutput;

      return res.json({
        status: "Success",
        user_email: userEmail,
        actual_output: stdout.trim(),
        expected_output: problemConfig.expectedOutput,
        pass: passed,
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

export default router;
