package main

import (
	"log"
	"net/http"
	"time"

	"github.com/spaceavocado/apidoc/example/handler/person"
	"github.com/gorilla/mux"
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

	log.Fatal(srv.ListenAndServe())
}
