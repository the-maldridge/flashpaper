package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/hashicorp/go-hclog"

	"github.com/the-maldridge/flashpaper/pkg/storage"
	"github.com/the-maldridge/flashpaper/pkg/web"
)

func main() {
	loglevel := os.Getenv("LOG_LEVEL")
	if loglevel == "" {
		loglevel = "INFO"
	}

	appLogger := hclog.New(&hclog.LoggerOptions{
		Name:  "flashpaper",
		Level: hclog.LevelFromString(loglevel),
	})
	appLogger.Info("Initializing")

	ws, err := web.New(appLogger)
	if err != nil {
		appLogger.Error("Error initializing webserver", "error", err)
		return
	}

	st, err := storage.NewRedis(appLogger)
	if err != nil {
		appLogger.Error("Error connecting to storage", "error", err)
		return
	}

	ws.SetStorage(st)
	go ws.Serve(":8080")

	sigs := make(chan os.Signal, 1)
	done := make(chan struct{}, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		appLogger.Info("Shutting down")

		ws.Shutdown()
		close(done)
	}()

	<-done
	appLogger.Info("Goodbye!")
}
