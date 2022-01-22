package req

import (
	"github.com/imroc/req/v2/internal/util"
	"net/url"
	"strings"
)

type (
	// RequestMiddleware type is for request middleware, called before a request is sent
	RequestMiddleware func(*Client, *Request) error
)

func parseRequestURL(c *Client, r *Request) error {
	if len(r.PathParams) > 0 {
		for p, v := range r.PathParams {
			r.URL = strings.Replace(r.URL, "{"+p+"}", url.PathEscape(v), -1)
		}
	}
	if len(c.PathParams) > 0 {
		for p, v := range c.PathParams {
			r.URL = strings.Replace(r.URL, "{"+p+"}", url.PathEscape(v), -1)
		}
	}

	// Parsing request URL
	reqURL, err := url.Parse(r.URL)
	if err != nil {
		return err
	}

	// If Request.URL is relative path then added c.HostURL into
	// the request URL otherwise Request.URL will be used as-is
	if !reqURL.IsAbs() {
		r.URL = reqURL.String()
		if len(r.URL) > 0 && r.URL[0] != '/' {
			r.URL = "/" + r.URL
		}

		reqURL, err = url.Parse(c.HostURL + r.URL)
		if err != nil {
			return err
		}
	}

	// GH #407 && #318
	if reqURL.Scheme == "" && len(c.scheme) > 0 {
		reqURL.Scheme = c.scheme
	}

	// Adding Query Param
	query := make(url.Values)
	for k, v := range c.QueryParams {
		for _, iv := range v {
			query.Add(k, iv)
		}
	}

	for k, v := range r.QueryParams {
		// remove query param from client level by key
		// since overrides happens for that key in the request
		query.Del(k)

		for _, iv := range v {
			query.Add(k, iv)
		}
	}

	// Preserve query string order partially.
	// Since not feasible in `SetQuery*` resty methods, because
	// standard package `url.Encode(...)` sorts the query params
	// alphabetically
	if len(query) > 0 {
		if util.IsStringEmpty(reqURL.RawQuery) {
			reqURL.RawQuery = query.Encode()
		} else {
			reqURL.RawQuery = reqURL.RawQuery + "&" + query.Encode()
		}
	}

	r.URL = reqURL.String()

	return nil
}

