package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/spaceavocado/apidoc/example/handler/person"
)

// @title Example API
// @desc Lorem ipsum dolor sit amet, consectetur adipiscing elit.
// Nullam rhoncus magna nunc, in faucibus metus pulvinar et.
// Mauris pellentesque enim justo.
// @terms https://domain.go/docs/api/terms
//
// @contact.name API Support
// @contact.url https://domain.go/contact
// @contact.email support@domain.go
//
// @lic.name Apache 2.0
// @lic.url https://www.apache.org/licenses/LICENSE-2.0.html
//
// @ver 1.0
// @server https://auth.domain.go/v3 Production API
// @server https://auth.dev.domain.go/v3 Development API
func main() {
	r := mux.NewRouter()
	srv := &http.Server{
		Handler:      r,
		Addr:         ":8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	// Person handlers
	person.Handlers(r)

	// Subrouter example:
	// @router animals
	sa := r.PathPrefix("/animal").Subrouter()

	// @router cats
	// @subrouter animals
	sc := sa.PathPrefix("/cat").Subrouter()

	// @produce plain/text
	// @success 200 {string} OK
	// @subrouter cats
	sc.HandleFunc("/list", CatList).Methods("GET")

	log.Fatal(srv.ListenAndServe())
}

// CatList request
func CatList(w http.ResponseWriter, r *http.Request) {}
