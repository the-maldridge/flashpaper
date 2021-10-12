package main

import (
	"context"

	"github.com/buraksezer/olric"
	"github.com/buraksezer/olric/config"
	"github.com/hashicorp/go-hclog"

	"github.com/the-maldridge/flashpaper/pkg/web"
)

func main() {
	appLogger := hclog.New(&hclog.LoggerOptions{
		Name:  "flashpaper",
		Level: hclog.LevelFromString("DEBUG"),
	})
	appLogger.Info("Initializing")

	ws, err := web.New(appLogger)
	if err != nil {
		appLogger.Error("Error initializing webserver", "error", err)
		return
	}

	cfg := config.New("local")
	cfg.Logger = appLogger.Named("olric").StandardLogger(&hclog.StandardLoggerOptions{InferLevels: true})

	ctx, cancel := context.WithCancel(context.Background())
	cfg.Started = func() {
		defer cancel()
		appLogger.Info("Olric has initialized")
	}

	db, err := olric.New(cfg)
	if err != nil {
		appLogger.Error("Error initializing storage", "error", err)
		return
	}

	go func() {
		if err := db.Start(); err != nil {
			appLogger.Error("Fatal error initializing olric", "error", err)
		}
	}()

	appLogger.Info("Waiting for early init to complete")
	<-ctx.Done()
	appLogger.Info("Early init complete")

	dm, err := db.NewDMap("storage")
	if err != nil {
		appLogger.Error("Error creating storage bucket", "error", err)
		return
	}
	ws.SetStorage(dm)

	ws.Serve(":8080")
}
