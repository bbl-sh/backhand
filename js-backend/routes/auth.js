import express from "express";
import PocketBase from "pocketbase";

const router = express.Router();

// Reuse the same PocketBase instance (imported from config/pocketbase.js)
const pb = new PocketBase("http://13.232.198.0:8890");

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
