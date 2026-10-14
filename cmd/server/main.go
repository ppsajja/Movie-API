// Local dev server: go run ./cmd/server  (reads .env if present)
package main

import (
	"bufio"
	"log"
	"net/http"
	"os"
	"strings"

	handler "movie-api/api"
)

func loadDotEnv() {
	f, err := os.Open(".env")
	if err != nil {
		return
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if ok && os.Getenv(k) == "" {
			os.Setenv(k, strings.Trim(v, `"`))
		}
	}
}

func main() {
	loadDotEnv()
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	http.HandleFunc("/", handler.Handler)
	log.Println("listening on http://localhost:" + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
