package rest

import "net/http"

func HandlerRestHttp() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello from http rest!"))
	}
}
