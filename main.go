package main

import (
	"log"
	"net/http"

	"github.com/iamjoona/chippy/api"
)

func main() {
	const filepathRoot = "."
	const port = "8080"

	apiCfg := &api.ApiConfig{}

	mux := http.NewServeMux()
	fsHandler := apiCfg.MiddlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot))))
	mux.Handle("/app/", fsHandler)
	mux.HandleFunc("GET /api/healthz", api.HandlerReadiness)
	mux.HandleFunc("GET /api/metrics", apiCfg.HandlerMetrics)
	mux.HandleFunc("POST /api/reset", apiCfg.HandlerReset)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())
}
