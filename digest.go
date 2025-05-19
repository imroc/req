package req

import (
	"net/http"

	"github.com/icholy/digest"
	"github.com/imroc/req/v3/internal/header"
)

// create response middleware for http digest authentication.
func handleDigestAuthFunc(username, password string) ResponseMiddleware {
	return func(client *Client, resp *Response) error {
		if resp.Err != nil || resp.StatusCode != http.StatusUnauthorized {
			return nil
		}
		auth, err := createDigestAuth(resp.Request.RawRequest, resp.Response, username, password)
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
		resp.Response, err = client.httpClient.Do(&req)
		return err
	}
}

func createDigestAuth(req *http.Request, resp *http.Response, username, password string) (auth string, err error) {
	chal, err := digest.FindChallenge(resp.Header)
	if err != nil {
		return "", err
	}
	cred, err := digest.Digest(chal, digest.Options{
		Username: username,
		Password: password,
		Method:   req.Method,
		URI:      req.URL.RequestURI(),
		GetBody:  req.GetBody,
		Count:    1,
	})
	if err != nil {
		return "", err
	}
	return cred.String(), nil
}
