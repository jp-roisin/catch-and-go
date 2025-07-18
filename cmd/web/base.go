package web

import (
	"log"
	"net/http"
)

func BaseWebHandler(w http.ResponseWriter, r *http.Request) {
	locale := "fr"

	component := Base(locale)
	err := component.Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Fatalf("Error rendering in HelloWebHandler: %e", err)
	}
}
