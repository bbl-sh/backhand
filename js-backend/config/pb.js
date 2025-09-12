import PocketBase from "pocketbase";
import "dotenv/config"; // ensure this runs before reading env

let pbInstance = null;

export async function getPB() {
  if (!pbInstance) {
    pbInstance = new PocketBase("http://13.232.198.0:8890/");
    pbInstance.autoCancellation(false);

    try {
      const email = (process.env.POCKETBASE_ADMIN_EMAIL || "").trim();
      const password = (process.env.POCKETBASE_ADMIN_PASSWORD || "").trim();

      await pbInstance
        .collection("_superusers")
        .authWithPassword(email, password, { autoRefreshThreshold: 30 * 60 });

      console.log("✅ PocketBase superuser authenticated.");
    } catch (error) {
      console.error("❌ Superuser authentication failed:", error);
    }
  }
  return pbInstance;
}
