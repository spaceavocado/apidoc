package common

// Address struct
type Address struct {
	Street  string `json:"street"`
	City    string `json:"city"`
	ZIP     string `json:"zip"`
	State   string `json:"state"`
	Country string `json:"country"`
}
