package handler

import (
	"log"
	"net/http"

	"github.com/faisalhardin/auth-vessel/internal/library/util/requestinfo"
)

func Handler(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// initial httpDone
		httpDone := make(chan bool)

		ctx := requestinfo.SetRequestContext(r.Context(), r)
		r = r.WithContext(ctx)

		// running background serve http
		panicCh := make(chan interface{}, 1)

		go func() {
			defer func() {
				if rec := recover(); rec != nil {
					log.Println("panic occurred:", rec)
					panicCh <- rec
				}
			}()

			next.ServeHTTP(w, r)
			httpDone <- true
		}()

		// get selection from context done data
		select {
		case <-panicCh:
			w.Header().Add("Content-Type", "application/json")
			w.Write([]byte(`{"error_messages":[{"error_name":"internal_server_error","error_description":"The server is unable to complete your request"}]}`))
		case <-ctx.Done():

			w.Write([]byte(`{"error_message": ["Process timeout"]}`))
		case <-httpDone:
		}
	})

}
