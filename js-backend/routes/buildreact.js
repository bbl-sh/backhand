// backend/routes/challenge.js
import express from "express";
import multer from "multer";
import { ProblemConfigs } from "../ProblemsConfigs/Challenges/buildreact/ProblemConfigs.js";
import { runDocker } from "../utils/docker.js";
import { verifyAuth } from "../middleware/verifyAuth.js";
import { fileURLToPath } from "url";
import path from "path";
import { getPB } from "../config/pb.js";

const router = express.Router();
const upload = multer({ dest: "uploads/" });

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

router.post(
  "/buildreact",
  upload.single("code"),
  verifyAuth,
  async (req, res) => {
    try {
      const problemId = req.body.problemId;
      const challengeId = req.body.ChallengeId || req.body.learnId;
      const challengeName = req.body.ChallengeName || req.body.learnName;

      console.log("Received challenge request:", {
        challengeId,
        challengeName,
        problemId,
      });

      const problemConfig = ProblemConfigs[problemId];
      if (!problemConfig) throw new Error("Invalid problemId");

      if (!req.file) throw new Error("No file uploaded");

      runDocker(
        problemConfig,
        req.file.path,
        __dirname,
        async (err, result) => {
          let responsePayload;

          if (err) {
            console.error("Docker execution error:", err.details);
            responsePayload = {
              status: "Not Accepted",
              message: err.stderr || err.details || "Compilation/Runtime Error",
              user_email: req.user.email,
              actual_output: "",
              expected_output: problemConfig.expectedOutput,
              pass: false,
              challengeId,
              challengeName,
              problemId,
              timestamp: new Date().toISOString(),
            };
          } else {
            responsePayload = {
              status: result.pass ? "Accepted" : "Not Accepted",
              message: result.pass ? "Output matched" : "Output did not match",
              user_email: req.user.email,
              actual_output: result.actual_output,
              expected_output: result.expected_output,
              pass: result.pass,
              challengeId,
              challengeName,
              problemId,
              timestamp: new Date().toISOString(),
            };
          }

          // === Update ChallengeProgress collection ===
          try {
            const pb = await getPB();
            const query = `email="${req.user.email}" && ChallengeName="${challengeName}"`;

            try {
              const existing = await pb
                .collection("ChallengeProgress")
                .getFirstListItem(query);

              // Update ChallengeName if needed
              await pb.collection("ChallengeProgress").update(existing.id, {
                email: req.user.email,
                ChallengeId: challengeId,
                ChallengeName: challengeName,
              });
            } catch (lookupErr) {
              if (lookupErr.status === 404) {
                // Create new record
                await pb.collection("ChallengeProgress").create({
                  email: req.user.email,
                  ChallengeId: challengeId,
                  ChallengeName: challengeName,
                });
              } else {
                console.error(
                  "❌ ChallengeProgress update failed:",
                  lookupErr.message,
                );
              }
            }
          } catch (dbErr) {
            console.error("❌ Error connecting to PocketBase:", dbErr.message);
          }

          return res.status(200).json(responsePayload);
        },
      );
    } catch (err) {
      console.error("Error in challenge handler:", err.message);
      return res.status(200).json({
        status: "Not Accepted",
        message: err.message,
        user_email: req.user?.email || null,
        actual_output: "",
        expected_output: null,
        pass: false,
        challengeId: req.body.challengeId || req.body.learnId || null,
        challengeName: req.body.challengeName || req.body.learnName || null,
        problemId: req.body.problemId || null,
        timestamp: new Date().toISOString(),
      });
    }
  },
);

export default router;
