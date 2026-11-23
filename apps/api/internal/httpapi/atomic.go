package httpapi

import (
	"errors"
	"net/http"
	"net/http/httptest"

	"github.com/savisaluwadana/Advance-HIRS-System/apps/api/internal/store"
)

var errMutationRejected = errors.New("mutation rejected")

// Buffer the JSON response until both the mutation and its audit event commit.
// Existing store transactions become savepoints on the request transaction.
func (s *Server) atomic(handler func(*Server, http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response := httptest.NewRecorder()
		err := s.store.InTransaction(r.Context(), func(scoped *store.Store) error {
			server := *s
			server.store = scoped
			handler(&server, response, r)
			if response.Code >= 400 {
				return errMutationRejected
			}
			return nil
		})
		if err != nil && !errors.Is(err, errMutationRejected) {
			writeError(w, http.StatusServiceUnavailable, "transaction_failed", "could not commit this change")
			return
		}
		for key, values := range response.Header() { w.Header()[key] = values }
		w.WriteHeader(response.Code)
		_, _ = w.Write(response.Body.Bytes())
	}
}
