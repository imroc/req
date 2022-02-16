package req

import "testing"

type User struct {
	Name string `json:"name" xml:"name"`
}

func TestUnmarshalJson(t *testing.T) {
	var user User
	resp, err := tc().R().Get("/json")
	assertSuccess(t, resp, err)
	err = resp.UnmarshalJson(&user)
	assertError(t, err)
	assertEqual(t, "roc", user.Name)
}

func TestUnmarshalXml(t *testing.T) {
	var user User
	resp, err := tc().R().Get("/xml")
	assertSuccess(t, resp, err)
	err = resp.UnmarshalXml(&user)
	assertError(t, err)
	assertEqual(t, "roc", user.Name)
}

func TestUnmarshal(t *testing.T) {
	var user User
	resp, err := tc().R().Get("/xml")
	assertSuccess(t, resp, err)
	err = resp.Unmarshal(&user)
	assertError(t, err)
	assertEqual(t, "roc", user.Name)
}
