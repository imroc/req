package req

import (
	"testing"
)

type User struct {
	Name string `json:"name" xml:"name"`
}

type Message struct {
	Message string `json:"message"`
}

func TestUnmarshalJson(t *testing.T) {
	var user User
	resp, err := tc().R().Get("/json")
	assertSuccess(t, resp, err)
	err = resp.UnmarshalJson(&user)
	assertNoError(t, err)
	assertEqual(t, "roc", user.Name)
}

func TestUnmarshalXml(t *testing.T) {
	var user User
	resp, err := tc().R().Get("/xml")
	assertSuccess(t, resp, err)
	err = resp.UnmarshalXml(&user)
	assertNoError(t, err)
	assertEqual(t, "roc", user.Name)
}

func TestUnmarshal(t *testing.T) {
	var user User
	resp, err := tc().R().Get("/xml")
	assertSuccess(t, resp, err)
	err = resp.Unmarshal(&user)
	assertNoError(t, err)
	assertEqual(t, "roc", user.Name)
}

func TestResponseResult(t *testing.T) {
	resp, _ := tc().R().SetResult(&User{}).Get("/json")
	user, ok := resp.Result().(*User)
	if !ok {
		t.Fatal("Response.Result() should return *User")
	}
	assertEqual(t, "roc", user.Name)

	assertEqual(t, true, resp.TotalTime() > 0)
	assertEqual(t, false, resp.ReceivedAt().IsZero())
}

func TestResponseError(t *testing.T) {
	resp, _ := tc().R().SetError(&Message{}).Get("/json?error=yes")
	msg, ok := resp.Error().(*Message)
	if !ok {
		t.Fatal("Response.Error() should return *Message")
	}
	assertEqual(t, "not allowed", msg.Message)
}
