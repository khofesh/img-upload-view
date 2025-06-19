package middlewares

import (
	"net/http"
)

// EnableCORS is a middleware to enable CORS
func (m *Middlewares[T]) EnableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Origin")

		w.Header().Add("Vary", "Access-Control-Request-Method")

		origin := r.Header.Get("Origin")

		if origin != "" {
			// Loop through the list of trusted origins, checking to see if the request
			// origin exactly matches one of them. If there are no trusted origins, then
			// the loop won't be iterated.
			for i := range m.trustedOrigins {
				if origin == m.trustedOrigins[i] {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Set("Access-Control-Allow-Credentials", "true") // for httpOnly cookie

					// Check if the request has the HTTP method OPTIONS and contains the
					// "Access-Control-Request-Method" header. If it does, then we treat
					// it as a preflight request.
					if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
						// Set the necessary preflight response headers
						w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
						w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization")

						w.WriteHeader(http.StatusOK)
						return
					}
					break
				}
			}
		}

		next.ServeHTTP(w, r)
	})
}
