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
    <script src="https://cdn.redoc.ly/redoc/latest/bundles/redoc.standalone.js"></script>
  </head>
  <body>
    <div id="redoc-container"></div>
    <script>
      Redoc.init('/openapi.yaml', {
        scrollYOffset: 50,
      }, document.getElementById('redoc-container'));
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
