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

		parts := strings.Split(raw, ",")
		ids := make([]int, 0, len(parts))
		for _, part := range parts {
			id, err := strconv.Atoi(strings.TrimSpace(part))
			if err != nil {
				http.Error(w, "invalid id: "+part, http.StatusBadRequest)
				return
			}
			ids = append(ids, id)
		}

		if err := dm.Store().ValidateStartingTeam(ids); err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
