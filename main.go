package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"github.com/floriancartron/cloudrun-revision-tag-urlviewer/cloudrun"
	"github.com/floriancartron/cloudrun-revision-tag-urlviewer/utils"
)

var logger *slog.Logger

func getData(w http.ResponseWriter, r *http.Request) {
	maxRevisions, err := strconv.Atoi(os.Getenv("CRRTUV_MAX_REVISIONS"))
	if err != nil {
		maxRevisions = 100
	}
	data, err := cloudrun.GetCloudRunData(logger, os.Getenv("CRRTUV_PROJECT"), os.Getenv("CRRTUV_LOCATION"), os.Getenv("CRRTUV_IDENTIFYING_LABEL"), maxRevisions)
	response := map[string]interface{}{
		"data": nil,
	}
	if err == nil {
		response = map[string]interface{}{
			"data": data,
		}
	} else {
		logger.Error(fmt.Sprintf("Error getting Cloud Run data: %v", err))
	}

	// Convert response to JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func serveHTML(w http.ResponseWriter, r *http.Request) {
	html, err := os.ReadFile("index.html")
	if err != nil {
		fmt.Println("Error reading index.html:", err)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	// Respond with a 200 status code and a simple message
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "OK")
}

func main() {
	loglevel := os.Getenv("CRRTUV_LOG_LEVEL")
	if loglevel == "" {
		loglevel = "DEBUG"
	}
	logger = utils.NewLogger(loglevel)
	http.HandleFunc("/", serveHTML)
	http.HandleFunc("/data", getData)
	http.HandleFunc("/healthz", healthHandler)

	logger.Info("Server is running on port 8080...")
	err := http.ListenAndServe("0.0.0.0:8080", nil)
	if err != nil {
		logger.Error(fmt.Sprintf("Error starting server: %v", err))
	}
}
