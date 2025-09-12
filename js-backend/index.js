import express from "express";
import authRoutes from "./routes/auth.js";
import challengeRoutes from "./routes/challenge.js";

const app = express();
const port = 3000;

// Middleware
app.use(express.json());

// Routes
app.use("/", authRoutes);
app.use("/", challengeRoutes);

app.listen(port, () => {
  console.log(`Server running at http://localhost:${port}`);
});
