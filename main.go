package main

import (
	"cloudrun-revision-tag-urlviewer/cloudrun"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

func getData(w http.ResponseWriter, r *http.Request) {
	data, err := cloudrun.GetCloudRunData(os.Getenv("CRRTUV_PROJECT"), os.Getenv("CRRTUV_LOCATION"), os.Getenv("CRRTUV_IDENTIFYING_LABEL"))
	response := map[string]interface{}{
		"data": nil,
	}
	if err == nil {
		response = map[string]interface{}{
			"data": data,
		}
	}

	// Convert response to JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func serveHTML(w http.ResponseWriter, r *http.Request) {
	html, err := ioutil.ReadFile("index.html")
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
	http.HandleFunc("/", serveHTML)
	http.HandleFunc("/data", getData)
	http.HandleFunc("/healthz", healthHandler)

	fmt.Println("Server is running on port 8080...")
	err := http.ListenAndServe("0.0.0.0:8080", nil)
	if err != nil {
		fmt.Printf("Error starting server: %v\n", err)
	}
}
