package req

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"errors"
	"fmt"
	"hash"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/imroc/req/v3/internal/header"
)

var (
	errDigestBadChallenge    = errors.New("digest: challenge is bad")
	errDigestCharset         = errors.New("digest: unsupported charset")
	errDigestAlgNotSupported = errors.New("digest: algorithm is not supported")
	errDigestQopNotSupported = errors.New("digest: no supported qop in list")
)

var hashFuncs = map[string]func() hash.Hash{
	"":                 md5.New,
	"MD5":              md5.New,
	"MD5-sess":         md5.New,
	"SHA-256":          sha256.New,
	"SHA-256-sess":     sha256.New,
	"SHA-512-256":      sha512.New,
	"SHA-512-256-sess": sha512.New,
}

// create response middleware for http digest authentication.
func handleDigestAuthFunc(username, password string) ResponseMiddleware {
	return func(client *Client, resp *Response) error {
		if resp.Err != nil || resp.StatusCode != http.StatusUnauthorized {
			return nil
		}
		auth, err := createDigestAuth(resp.Response, username, password)
		if err != nil {
			return err
		}
		r := resp.Request
		req := *r.RawRequest
		if req.Body != nil {
			err = parseRequestBody(client, r) // re-setup body
			if err != nil {
				return err
			}
			if r.GetBody != nil {
				body, err := r.GetBody()
				if err != nil {
					return err
				}
				req.Body = body
				req.GetBody = r.GetBody
			}
		}
		if req.Header == nil {
			req.Header = make(http.Header)
		}
		req.Header.Set(header.Authorization, auth)
		resp.Response, err = client.GetTransport().RoundTrip(&req)
		return err
	}
}

func createDigestAuth(resp *http.Response, username, password string) (auth string, err error) {
	chal := resp.Header.Get(header.WwwAuthenticate)
	if chal == "" {
		return "", errDigestBadChallenge
	}

	c, err := parseChallenge(chal)
	if err != nil {
		return "", err
	}

	// Form credentials based on the challenge
	cr := newCredentials(resp.Request.URL.RequestURI(), resp.Request.Method, username, password, c)
	auth, err = cr.authorize()
	return
}

func newCredentials(digestURI, method, username, password string, c *challenge) *credentials {
	return &credentials{
		username:   username,
		userhash:   c.userhash,
		realm:      c.realm,
		nonce:      c.nonce,
		digestURI:  digestURI,
		algorithm:  c.algorithm,
		sessionAlg: strings.HasSuffix(c.algorithm, "-sess"),
		opaque:     c.opaque,
		messageQop: c.qop,
		nc:         0,
		method:     method,
		password:   password,
	}
}

type challenge struct {
	realm     string
	domain    string
	nonce     string
	opaque    string
	stale     string
	algorithm string
	qop       string
	userhash  string
}

func parseChallenge(input string) (*challenge, error) {
	if !strings.HasPrefix(input, "Digest ") {
		return nil, errDigestBadChallenge
	}
	c := &challenge{}
	c.algorithm = "MD5"
	re := regexp.MustCompile(`([\w]+)="?([^",]+)"?`)
	matches := re.FindAllStringSubmatch(input, -1)

	for _, match := range matches {
		switch match[1] {
		case "realm":
			c.realm = match[2]
		case "algorithm":
			c.algorithm = match[2]
		case "nonce":
			c.nonce = match[2]
		case "stale":
			c.stale = match[2]
		case "qop":
			for _, v := range strings.Split(match[2], ",") {
				v = strings.Trim(v, " ")
				if v == "auth" || v == "auth-int" {
					c.qop = "auth"
					break
				}
			}
		case "userhash":
			c.userhash = match[2]
		case "domain":
			c.domain = match[2]
		case "opaque":
			c.opaque = match[2]
		default:
			return nil, errDigestBadChallenge
		}
	}
	return c, nil
}

type credentials struct {
	username   string
	userhash   string
	realm      string
	nonce      string
	digestURI  string
	algorithm  string
	sessionAlg bool
	cNonce     string
	opaque     string
	messageQop string
	nc         int
	method     string
	password   string
}

func (c *credentials) authorize() (string, error) {
	if _, ok := hashFuncs[c.algorithm]; !ok {
		return "", errDigestAlgNotSupported
	}

	if err := c.validateQop(); err != nil {
		return "", err
	}

	resp, err := c.resp()
	if err != nil {
		return "", err
	}

	sl := make([]string, 0, 10)
	if c.userhash == "true" {
		// RFC 7616 3.4.4
		c.username = c.h(fmt.Sprintf("%s:%s", c.username, c.realm))
		sl = append(sl, fmt.Sprintf(`userhash=%s`, c.userhash))
	}
	sl = append(sl, fmt.Sprintf(`username="%s"`, c.username))
	sl = append(sl, fmt.Sprintf(`realm="%s"`, c.realm))
	sl = append(sl, fmt.Sprintf(`nonce="%s"`, c.nonce))
	sl = append(sl, fmt.Sprintf(`uri="%s"`, c.digestURI))
	sl = append(sl, fmt.Sprintf(`response="%s"`, resp))
	sl = append(sl, fmt.Sprintf(`algorithm=%s`, c.algorithm))
	if c.opaque != "" {
		sl = append(sl, fmt.Sprintf(`opaque="%s"`, c.opaque))
	}
	if c.messageQop != "" {
		sl = append(sl, fmt.Sprintf("qop=%s", c.messageQop))
		sl = append(sl, fmt.Sprintf("nc=%08x", c.nc))
		sl = append(sl, fmt.Sprintf(`cnonce="%s"`, c.cNonce))
	}

	return fmt.Sprintf("Digest %s", strings.Join(sl, ", ")), nil
}

func (c *credentials) validateQop() error {
	// Currently only supporting auth quality of protection. TODO: add auth-int support
	if c.messageQop == "" {
		return nil
	}
	possibleQops := strings.Split(c.messageQop, ", ")
	var authSupport bool
	for _, qop := range possibleQops {
		if qop == "auth" {
			authSupport = true
			break
		}
	}
	if !authSupport {
		return errDigestQopNotSupported
	}

	return nil
}

func (c *credentials) h(data string) string {
	hfCtor := hashFuncs[c.algorithm]
	hf := hfCtor()
	_, _ = hf.Write([]byte(data)) // Hash.Write never returns an error
	return fmt.Sprintf("%x", hf.Sum(nil))
}

func (c *credentials) resp() (string, error) {
	c.nc++

	b := make([]byte, 16)
	_, err := io.ReadFull(rand.Reader, b)
	if err != nil {
		return "", err
	}
	c.cNonce = fmt.Sprintf("%x", b)[:32]

	ha1 := c.ha1()
	ha2 := c.ha2()
	if c.messageQop == "" {
		return c.h(fmt.Sprintf("%s:%s:%s", ha1, c.nonce, ha2)), nil
	}
	return c.kd(ha1, fmt.Sprintf("%s:%08x:%s:%s:%s",
		c.nonce, c.nc, c.cNonce, c.messageQop, ha2)), nil
}

func (c *credentials) kd(secret, data string) string {
	return c.h(fmt.Sprintf("%s:%s", secret, data))
}

// RFC 7616 3.4.2
func (c *credentials) ha1() string {
	ret := c.h(fmt.Sprintf("%s:%s:%s", c.username, c.realm, c.password))
	if c.sessionAlg {
		return c.h(fmt.Sprintf("%s:%s:%s", ret, c.nonce, c.cNonce))
	}

	return ret
}

// RFC 7616 3.4.3
func (c *credentials) ha2() string {
	// currently no auth-int support
	return c.h(fmt.Sprintf("%s:%s", c.method, c.digestURI))
}
