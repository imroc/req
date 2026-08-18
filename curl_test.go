package req

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/imroc/req/v3/internal/tests"
)

func assertCurlContains(t *testing.T, cmd, substr string) {
	t.Helper()
	if !strings.Contains(cmd, substr) {
		t.Errorf("%q is not included in %q", substr, cmd)
	}
}

func TestGenerateCurlCommandGet(t *testing.T) {
	r := C().R().
		SetHeader("Accept", "application/json").
		SetQueryParam("page", "2")
	r.Method = http.MethodGet
	r.SetURL("https://api.example.com/users")

	tests.AssertEqual(t,
		"curl 'https://api.example.com/users?page=2' -H 'Accept: application/json'",
		r.GenerateCurlCommand())
}

func TestGenerateCurlCommandPostJson(t *testing.T) {
	r := C().R().
		SetBearerAuthToken("secret-token").
		SetBodyJsonString(`{"name":"req"}`)
	r.Method = http.MethodPost
	r.SetURL("https://api.example.com/users")

	tests.AssertEqual(t,
		"curl -X POST 'https://api.example.com/users' -H 'Authorization: Bearer secret-token' -H 'Content-Type: application/json; charset=utf-8' -d '{\"name\":\"req\"}'",
		r.GenerateCurlCommand())
}

func TestGenerateCurlCommandFormData(t *testing.T) {
	r := C().R().
		SetFormData(map[string]string{"name": "req"})
	r.Method = http.MethodPost
	r.SetURL("https://api.example.com/login")

	cmd := r.GenerateCurlCommand()
	assertCurlContains(t, cmd, "-X POST")
	assertCurlContains(t, cmd, "-H 'Content-Type: application/x-www-form-urlencoded'")
	assertCurlContains(t, cmd, "-d 'name=req'")
}

func TestGenerateCurlCommandCookies(t *testing.T) {
	r := C().R().
		SetCookies(
			&http.Cookie{Name: "sid", Value: "abc"},
			&http.Cookie{Name: "theme", Value: "dark"},
		)
	r.Method = http.MethodGet
	r.SetURL("https://example.com/")

	assertCurlContains(t, r.GenerateCurlCommand(), "-H 'Cookie: sid=abc; theme=dark'")
}

func TestGenerateCurlCommandBasicAuth(t *testing.T) {
	r := C().R().SetBasicAuth("imroc", "123456")
	r.Method = http.MethodGet
	r.SetURL("https://example.com/")

	assertCurlContains(t, r.GenerateCurlCommand(), "-H 'Authorization: Basic aW1yb2M6MTIzNDU2'")
}

func TestGenerateCurlCommandShellEscape(t *testing.T) {
	r := C().R().SetHeader("X-Test", "a'b")
	r.Method = http.MethodGet
	r.SetURL("https://example.com/")

	assertCurlContains(t, r.GenerateCurlCommand(), `-H 'X-Test: a'\''b'`)
}

// TestGenerateCurlCommandDoesNotMutate verifies that generating the command
// before sending leaves the original request untouched, so it can still be
// sent afterwards without, for example, cookies being appended twice.
func TestGenerateCurlCommandDoesNotMutate(t *testing.T) {
	r := C().R().SetCookies(&http.Cookie{Name: "sid", Value: "abc"})
	r.Method = http.MethodGet
	r.SetURL("https://example.com/")

	_ = r.GenerateCurlCommand()
	_ = r.GenerateCurlCommand()

	tests.AssertEqual(t, 1, len(r.Cookies))
	tests.AssertEqual(t, true, r.URL == nil)
	tests.AssertEqual(t, true, r.RawRequest == nil)
}

// TestGenerateCurlCommandAfterSend verifies that the command reflects the
// request that was actually fired once it has been sent.
func TestGenerateCurlCommandAfterSend(t *testing.T) {
	r := tc().R().SetHeader("X-Foo", "bar")
	_, _ = r.Get("/")

	tests.AssertNotNil(t, r.RawRequest)
	cmd := r.GenerateCurlCommand()
	assertCurlContains(t, cmd, "curl")
	assertCurlContains(t, cmd, "-H 'X-Foo: bar'")
	assertCurlContains(t, cmd, getTestServerURL())
}

func TestGenerateCurlCommandNilOrEmpty(t *testing.T) {
	var r *Request
	tests.AssertEqual(t, "", r.GenerateCurlCommand())

	empty := &Request{}
	tests.AssertEqual(t, "", empty.GenerateCurlCommand())
}

func TestGenerateCurlCommandHostOverride(t *testing.T) {
	u, err := url.Parse("https://1.2.3.4/health")
	tests.AssertNoError(t, err)
	req := &http.Request{
		Method: http.MethodGet,
		Host:   "internal.example.com",
		URL:    u,
		Header: make(http.Header),
	}

	cmd := buildCurlCommand(req, nil)
	if !strings.Contains(cmd, "-H 'Host: internal.example.com'") {
		t.Fatalf("expected Host header override in curl command, got: %s", cmd)
	}
}
