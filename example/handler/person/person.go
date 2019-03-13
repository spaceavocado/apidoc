package person

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/spaceavocado/apidoc/example/request"
	"github.com/spaceavocado/apidoc/example/response"
	// Common respose, structs referenced in the API documentation
	_ "github.com/spaceavocado/apidoc/example/common"
)

// Person response model
type Person struct {
	Name string `json:"fullname" required:"true"`
	// User's Profile
	Detail Detail `json:"profile"`
}

// Detail of the user
type Detail struct {
	Age    int64  `json:"age"`
	Status Status `json:"status" apitype:"int"`
}

// Status of the person
type Status int

// Handlers for this resource / API section
func Handlers(r *mux.Router) {
	// GetPerson handler
	// @summary Person Address
	// @desc Get person address.
	// @id person-address
	// @tag Person
	// @produce json
	// @success 200 {object} common.Address OK
	// @failure 500 {string} Internal Server Error
	r.HandleFunc("/person/{id:[0-9]+}/address", GetAddress).Methods("GET")

	// GetPerson handler
	// @summary Person
	// @desc Get person by ID.
	// @id person
	// @tag Person
	// @produce json
	// @success 200 {object} Person OK
	// @failure 500 {string} Internal Server Error
	r.HandleFunc("/person/{id:[0-9]+}", GetPerson).Methods("GET")

	r.HandleFunc("/person", CreatePerson).Methods("PUT")
}

// GetAddress request
func GetAddress(w http.ResponseWriter, r *http.Request) {
	// Not implemented
}

// GetPerson request
func GetPerson(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fmt.Printf("Get person with %s ID", vars["id"])

	response.JSON(w, 200, Person{
		Name: "Peter Williams",
		Detail: Detail{
			Age: 55,
		},
	})
}

// CreatePerson handler
// @summary Create
// @desc Create a new Person
// @id create-person
// @tag Person
// @accept json
// @produce json
// @body Person
// @success 200 {string} OK
// @fwrap response.Error error
// @failure 500 {object} response.APIError Internal Server Error
// @router /person [put]
func CreatePerson(w http.ResponseWriter, r *http.Request) {
	var model Person
	err := request.ParseJSONBody(r.Body, &model)
	if err != nil {
		response.APIResponseError(w, response.APIError{
			Code:    1,
			Message: "error message",
		})
		return
	}

	fmt.Printf("Create a new person '%s'", model.Name)
}
