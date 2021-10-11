package main

import (
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

	ws.Serve(":8080")
}
