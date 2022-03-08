package socks

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestReply(t *testing.T) {
	for i := 0; i < 9; i++ {
		s := Reply(i).String()
		if strings.Contains(s, "unknown") {
			t.Errorf("resply code [%d] should not unknown", i)
		}
	}
	s := Reply(9).String()
	if !strings.Contains(s, "unknown") {
		t.Errorf("resply code [%d] should unknown", 9)
	}
}

func assertNoError(t *testing.T, err error) {
	if err != nil {
		t.Errorf("Error occurred [%v]", err)
	}
}

func assertErrorContains(t *testing.T, err error, s string) {
	if err == nil {
		t.Error("err is nil")
		return
	}
	if !strings.Contains(err.Error(), s) {
		t.Errorf("%q is not included in error %q", s, err.Error())
	}
}

func TestAuthenticate(t *testing.T) {
	auth := &UsernamePassword{
		Username: "imroc",
		Password: "123456",
	}
	buf := bytes.NewBuffer([]byte{byte(0x01), byte(0x00)})
	err := auth.Authenticate(context.Background(), buf, AuthMethodUsernamePassword)
	assertNoError(t, err)
	auth.Username = "this is a very long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long name"
	err = auth.Authenticate(context.Background(), buf, AuthMethodUsernamePassword)
	assertErrorContains(t, err, "invalid")

	auth.Username = "imroc"
	buf = bytes.NewBuffer([]byte{byte(0x03), byte(0x00)})
	err = auth.Authenticate(context.Background(), buf, AuthMethodUsernamePassword)
	assertErrorContains(t, err, "invalid username/password version")

	buf = bytes.NewBuffer([]byte{byte(0x01), byte(0x02)})
	err = auth.Authenticate(context.Background(), buf, AuthMethodUsernamePassword)
	assertErrorContains(t, err, "authentication failed")

	err = auth.Authenticate(context.Background(), buf, AuthMethodNoAcceptableMethods)
	assertErrorContains(t, err, "unsupported authentication method")

}
