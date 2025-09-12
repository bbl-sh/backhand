import jwt from "jsonwebtoken";

export function decodeToken(token) {
  try {
    return jwt.decode(token);
  } catch (err) {
    console.error("JWT decode error:", err.message);
    return null;
  }
}

// Optional: If you want to export as default too
export default decodeToken;
