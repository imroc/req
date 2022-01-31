package req

import (
	"net/http"
	"testing"
)

func TestGet(t *testing.T) {
	ts := createGetServer(t)
	defer ts.Close()

	c := tc()
	resp, err := c.R().Get(ts.URL)
	assertResponse(t, resp, err)
	assertEqual(t, "TestGet: text response", resp.String())

	resp, err = c.R().Get(ts.URL + "/no-content")
	assertResponse(t, resp, err)
	assertEqual(t, "", resp.String())

	resp, err = c.R().Get(ts.URL + "/json")
	assertResponse(t, resp, err)
	assertEqual(t, `{"TestGet": "JSON response"}`, resp.String())
	assertEqual(t, resp.GetContentType(), "application/json")

	resp, err = c.R().Get(ts.URL + "/json-invalid")
	assertResponse(t, resp, err)
	assertEqual(t, `TestGet: Invalid JSON`, resp.String())
	assertEqual(t, resp.GetContentType(), "application/json")

	resp, err = c.R().Get(ts.URL + "/bad-request")
	assertStatus(t, resp, err, http.StatusBadRequest, "400 Bad Request")
}
