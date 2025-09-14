import express from "express";
import pb from "../config/pocketbase.js";

const router = express.Router();

/**
 * === Login Route ===
 * POST /login
 * Body: { email, password }
 */
router.post("/login", async (req, res) => {
  try {
    const { email, password } = req.body;

    if (!email || !password) {
      throw new Error("Email and password are required");
    }

    const authData = await pb
      .collection("users")
      .authWithPassword(email, password);

    // FIX: Changed response structure to match what the client script (setup.sh) expects.
    // - Use `success: true` for successful authentication.
    return res.json({
      success: true,
      email: authData.record.email,
      id: authData.record.id,
      token: authData.token,
    });
  } catch (err) {
    console.error("Login error:", err.message);

    // FIX: Changed the error response structure for consistency with the client script.
    // - Use `success: false` for failed authentication.
    // - Send the error message in an `error` field.
    return res.status(400).json({
      success: false,
      error: err.message,
    });
  }
});

export default router;
