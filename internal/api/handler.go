package api

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/vipulvpatil/whatfpl/internal/fpl"
)

func NewHandler(dm *fpl.DataManager) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /players", handlePlayers(dm))
	return mux
}

func handlePlayers(dm *fpl.DataManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		raw := r.URL.Query().Get("ids")
		if raw == "" {
			http.Error(w, "missing ids", http.StatusBadRequest)
			return
		}

		for _, part := range strings.Split(raw, ",") {
			if _, err := strconv.Atoi(strings.TrimSpace(part)); err != nil {
				http.Error(w, "invalid id: "+part, http.StatusBadRequest)
				return
			}
		}

		w.WriteHeader(http.StatusOK)
	}
}
