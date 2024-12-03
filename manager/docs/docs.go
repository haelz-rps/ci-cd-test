package docs

import (
	_ "embed"
	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger"
	"net/http"
)

//go:embed open-api-schema.yaml
var openAPIDoc []byte

func NewDocsHandler() http.Handler {
	r := chi.NewRouter()

	r.Get("/schema", func(w http.ResponseWriter, r *http.Request) {
		w.Write(openAPIDoc)
	})

	r.Get("/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/schema"),
	))

	return r
}
