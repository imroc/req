package socks

import (
	"bytes"
	"context"
	"github.com/imroc/req/v3/internal/tests"
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

func TestAuthenticate(t *testing.T) {
	auth := &UsernamePassword{
		Username: "imroc",
		Password: "123456",
	}
	buf := bytes.NewBuffer([]byte{byte(0x01), byte(0x00)})
	err := auth.Authenticate(context.Background(), buf, AuthMethodUsernamePassword)
	tests.AssertNoError(t, err)
	auth.Username = "this is a very long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long long name"
	err = auth.Authenticate(context.Background(), buf, AuthMethodUsernamePassword)
	tests.AssertErrorContains(t, err, "invalid")

	auth.Username = "imroc"
	buf = bytes.NewBuffer([]byte{byte(0x03), byte(0x00)})
	err = auth.Authenticate(context.Background(), buf, AuthMethodUsernamePassword)
	tests.AssertErrorContains(t, err, "invalid username/password version")

	buf = bytes.NewBuffer([]byte{byte(0x01), byte(0x02)})
	err = auth.Authenticate(context.Background(), buf, AuthMethodUsernamePassword)
	tests.AssertErrorContains(t, err, "authentication failed")

	err = auth.Authenticate(context.Background(), buf, AuthMethodNoAcceptableMethods)
	tests.AssertErrorContains(t, err, "unsupported authentication method")

}
