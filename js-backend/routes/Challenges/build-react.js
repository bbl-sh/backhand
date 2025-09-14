import express from "express";
import multer from "multer";
import os from "os"; // Import os module
import { ProblemConfigs } from "../../ProblemsConfigs/Challenges/buildreact/ProblemConfigs.js";
import { runDocker } from "../../utils/docker.js";
import { verifyAuth } from "../../middleware/verifyAuth.js";
import { getPB } from "../../config/pb.js";

const router = express.Router();
// Use the OS's temporary directory for uploads. It's safer.
const upload = multer({ dest: os.tmpdir() });

router.post(
  "/buildreact",
  upload.single("code"),
  verifyAuth,
  async (req, res) => {
    // --- Basic Validation ---
    if (!req.file) {
      return res.status(400).json({ message: "No code file was uploaded." });
    }

    const { ChallengeId, learnId, ChallengeName, learnName } = req.body;
    const challengeId = parseInt(ChallengeId || learnId, 10);
    const challengeName = ChallengeName || learnName;

    // The problemId is now derived directly from the challengeId
    const problemId = challengeId;
    const problemConfig = ProblemConfigs[problemId];

    if (!problemConfig) {
      return res.status(400).json({ message: "Invalid problemId." });
    }

    // --- Pre-check for existing completion in PocketBase ---
    try {
      const pb = await getPB();
      const query = `email="${req.user.email}" && ChallengeName="${challengeName}"`;
      const existing = await pb
        .collection("ChallengeProgress")
        .getFirstListItem(query)
        .catch((err) => {
          if (err.status === 404) return null; // Not found is okay, means it's their first attempt
          throw err; // Re-throw other errors
        });

      if (existing && challengeId <= existing.ChallengeId) {
        return res.status(200).json({
          status: "Skipped",
          message: "This challenge has already been completed.",
          pass: true, // Mark as pass to avoid frontend errors
          user_email: req.user.email,
          challengeId: challengeId,
          challengeName: challengeName,
          problemId: problemId,
          timestamp: new Date().toISOString(),
        });
      }
    } catch (dbError) {
      console.error("❌ PocketBase pre-check failed:", dbError.message);
      // We will still proceed, but log that the check failed.
    }

    // --- Main Logic ---
    try {
      const result = await runDocker(problemConfig, req.file.path);

      const responsePayload = {
        status: result.pass ? "Accepted" : "Not Accepted",
        message: result.pass ? "Output matched" : "Output did not match",
        user_email: req.user.email,
        pass: result.pass,
        actual_output: result.actual_output,
        expected_output: result.expected_output,
        challengeId: challengeId,
        challengeName: challengeName,
        problemId: problemId,
        timestamp: new Date().toISOString(),
      };

      // --- Database Update on Success ---
      if (result.pass) {
        try {
          const pb = await getPB();
          const query = `email="${req.user.email}" && ChallengeName="${challengeName}"`;

          try {
            const existing = await pb
              .collection("ChallengeProgress")
              .getFirstListItem(query);
            // Record exists, update it with the new, higher ChallengeId
            await pb.collection("ChallengeProgress").update(existing.id, {
              ChallengeId: challengeId,
            });
          } catch (lookupErr) {
            if (lookupErr.status === 404) {
              // No record exists, create a new one
              await pb.collection("ChallengeProgress").create({
                email: req.user.email,
                ChallengeId: challengeId,
                ChallengeName: challengeName,
              });
            } else {
              // Handle other potential errors during lookup
              console.error(
                "❌ ChallengeProgress update failed:",
                lookupErr.message,
              );
            }
          }
        } catch (dbError) {
          console.error("❌ PocketBase update failed:", dbError.message);
          // Don't block the response to the user if the DB fails
        }
      }

      return res.status(200).json(responsePayload);
    } catch (err) {
      // --- Error Handling from Docker ---
      console.error("Error during code execution:", err);
      return res.status(200).json({
        // Still send 200, but with Not Accepted status
        status: "Not Accepted",
        message:
          err.stderr ||
          err.message ||
          "An unknown error occurred during execution.",
        user_email: req.user.email,
        pass: false,
        actual_output: err.stdout || "",
        expected_output: problemConfig.expectedOutput,
        challengeId: challengeId,
        challengeName: challengeName,
        problemId: problemId,
        timestamp: new Date().toISOString(),
      });
    }
  },
);

export default router;
