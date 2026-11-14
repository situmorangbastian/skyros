package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func main() {
	cfg := viper.New()
	cfg.AutomaticEnv()
	cfg.SetConfigFile(".env")
	if err := cfg.ReadInConfig(); err != nil {
		log.Fatal("failed to read config file")
	}

	r := mux.NewRouter()

	r.HandleFunc("/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		openapiPath := filepath.Join(".", "openapi.yaml")
		if _, err := os.Stat(openapiPath); os.IsNotExist(err) {
			http.Error(w, "openapi.yaml not found", http.StatusNotFound)
			return
		}
		http.ServeFile(w, r, openapiPath)
	})

	const docsHTML = `<!DOCTYPE html>
<html>
  <head>
    <title>API Docs</title>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <style>
      body { margin: 0; padding: 0; }
    </style>
  </head>
  <body>
    <div id="app"></div>
    <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference@1.76.0/dist/browser/standalone.js"></script>
    <script>
      Scalar.createApiReference('#app', {
        url: '/openapi.yaml',
        layout: 'modern',
        defaultHttpClient: { targetKey: 'shell', clientKey: 'curl' },
      });
    </script>
  </body>
</html>`

	r.HandleFunc("/docs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")

		if _, err := io.WriteString(w, docsHTML); err != nil {
			log.Printf("failed to write response: %v", err)
		}
	})

	log.Info("server running on port: ", cfg.GetInt("PORT"))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", cfg.GetInt("PORT")), r))

}
