import express from "express";
import authRoutes from "./routes/auth.js";
import challengeRoutes from "./routes/challenge.js";
import challengesRoutes09 from "./routes/Challenges/build-react.js";
const app = express();
const port = 3000;

// Middleware
app.use(express.json());

// Routes
app.use("/", authRoutes);
app.use("/", challengeRoutes);

// Build your own react route
app.use("/", challengesRoutes09);

app.listen(port, () => {
  console.log(`Server running at http://localhost:${port}`);
});
