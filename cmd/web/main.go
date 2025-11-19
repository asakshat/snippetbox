package main

import (
	"log/slog"
	"net/http"
	"os"
)

type application struct {
	logger *slog.Logger
}

func main() {
    logger:= slog.New(slog.NewTextHandler(os.Stdout,nil))
	app := &application {
		logger: logger,
	}

    port:= ":4000"
	logger.Info("starting server","addr",port)

	err := http.ListenAndServe(port,app.routes())
    logger.Error(err.Error())
	os.Exit(1)
}
