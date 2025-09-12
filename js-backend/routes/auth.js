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

export default router;
