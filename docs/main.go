package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	cfg "github.com/spf13/viper"
)

func main() {
	cfg.SetConfigFile(".env")
	cfg.AutomaticEnv()
	cfg.ReadInConfig()

	r := mux.NewRouter()

	r.HandleFunc("/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		openapiPath := filepath.Join(".", "openapi.yaml")
		if _, err := os.Stat(openapiPath); os.IsNotExist(err) {
			http.Error(w, "openapi.yaml not found", http.StatusNotFound)
			return
		}
		http.ServeFile(w, r, openapiPath)
	})

	r.HandleFunc("/docs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`
<!DOCTYPE html>
<html>
  <head>
    <title>API Docs</title>
    <script src="https://cdn.redoc.ly/redoc/latest/bundles/redoc.standalone.js"></script>
  </head>
  <body>
    <div id="redoc-container"></div> <!-- ReDoc container -->
    <script>
      // Ensure the ReDoc spec is loaded properly
      Redoc.init('/openapi.yaml', {
        scrollYOffset: 50,
      }, document.getElementById('redoc-container'));
    </script>
  </body>
</html>`))
	})

	log.Info("server running on port: ", cfg.GetInt("PORT"))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", cfg.GetInt("PORT")), r))

}
