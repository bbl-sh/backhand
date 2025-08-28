package pb

import (
	"log"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/sdk"
)

// StartEmbedded starts PocketBase in a goroutine and returns an SDK client
// pointing to the embedded server (http://127.0.0.1:8090).
// The function returns after PocketBase has started serving.
func StartEmbedded() *sdk.Client {
	app := pocketbase.New()

	// Optional: serve static files from ./pb_public if needed
	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		se.Router.GET("/{path...}", apis.Static(nil, false))
		return se.Next()
	})

	// Start PocketBase in a goroutine (Start blocks).
	go func() {
		if err := app.Start(); err != nil {
			log.Fatalf("pocketbase start error: %v", err)
		}
	}()

	// Wait a short while for PB to come up (simple heuristic).
	// In production you might poll the health endpoint instead.
	time.Sleep(800 * time.Millisecond)

	// Create SDK client pointing to the embedded server
	client := sdk.NewClient("http://127.0.0.1:8090")

	return client
}
