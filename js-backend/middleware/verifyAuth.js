import pb from "../config/pocketbase.js";

export async function verifyAuth(req, res, next) {
  try {
    const authHeader = req.headers["authorization"];
    const userEmail = req.headers["x-user-email"];

    if (!authHeader || !userEmail) {
      return res.status(401).json({ error: "Missing auth headers" });
    }

    const token = authHeader.split(" ")[1];
    if (!token) {
      return res.status(401).json({ error: "Invalid Authorization header" });
    }

    // Save token & refresh via PocketBase
    pb.authStore.save(token, null);
    const authData = await pb.collection("users").authRefresh();

    if (!authData.record || authData.record.email !== userEmail) {
      return res
        .status(401)
        .json({ error: "User not found or email mismatch" });
    }

    req.user = authData.record; // attach user to request
    next();
  } catch (err) {
    return res
      .status(401)
      .json({ error: "Authentication failed", details: err.message });
  }
}
