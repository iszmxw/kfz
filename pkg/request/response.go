package request

import (
	"encoding/json"
	"errors"
	cmap "github.com/orcaman/concurrent-map"
	"io/ioutil"
	"net/http"
)

func ParseResponse(response *http.Response) (map[string]interface{}, http.Header, error) {
	if response.StatusCode != 200 {
		return nil, response.Header, errors.New(response.Status)
	}
	result := cmap.New().Items()
	body, err := ioutil.ReadAll(response.Body)
	if err == nil {
		err = json.Unmarshal(body, &result)
	}
	return result, response.Header, err
}
