package request

import (
	"encoding/json"
	"io"
	"io/ioutil"
)

// ParseJSONBody parses the request JSON body
// into the desired struct
func ParseJSONBody(body io.ReadCloser, o interface{}) error {
	raw, err := ioutil.ReadAll(body)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(raw), &o)
}
