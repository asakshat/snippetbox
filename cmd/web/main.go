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
    logger:= slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{AddSource: true,}))
	mux := http.NewServeMux()

	app := &application {
		logger: logger,
	}

	fileServer:= http.FileServer(http.Dir("./ui/static/"))
	mux.Handle("GET /static/", http.StripPrefix("/static", fileServer))

	mux.HandleFunc("GET /{$}" , app.home)
	mux.HandleFunc("GET /snippet/view/{id}", app.snippetView)
	mux.HandleFunc("GET /snippet/create", app.snippetCreate)
	mux.HandleFunc("POST /snippet/create", app.snippetCreatePost)
    port:= ":4000"
	logger.Info("starting server","addr",port)
	err := http.ListenAndServe(port,mux)

    logger.Error(err.Error())
	os.Exit(1)
}
