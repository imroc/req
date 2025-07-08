package github

import (
	"context"
	"fmt"
	"github.com/0xobjc/req/v3"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"strconv"
	"strings"
)

// Client is the go client for GitHub API.
type Client struct {
	*req.Client
}

// APIError represents the error message that GitHub API returns.
// GitHub API doc: https://docs.github.com/en/rest/overview/resources-in-the-rest-api#client-errors
type APIError struct {
	Message          string `json:"message"`
	DocumentationUrl string `json:"documentation_url,omitempty"`
	Errors           []struct {
		Resource string `json:"resource"`
		Field    string `json:"field"`
		Code     string `json:"code"`
	} `json:"errors,omitempty"`
}

// Error convert APIError to a human readable error and return.
func (e *APIError) Error() string {
	msg := fmt.Sprintf("API error: %s", e.Message)
	if e.DocumentationUrl != "" {
		return fmt.Sprintf("%s (see doc %s)", msg, e.DocumentationUrl)
	}
	if len(e.Errors) == 0 {
		return msg
	}
	errs := []string{}
	for _, err := range e.Errors {
		errs = append(errs, fmt.Sprintf("resource:%s field:%s code:%s", err.Resource, err.Field, err.Code))
	}
	return fmt.Sprintf("%s (%s)", msg, strings.Join(errs, " | "))
}

// NewClient create a GitHub client.
func NewClient() *Client {
	c := req.C().
		// All GitHub API requests need this header.
		SetCommonHeader("Accept", "application/vnd.github.v3+json").
		// All GitHub API requests use the same base URL.
		SetBaseURL("https://api.github.com").
		// Enable dump at the request-level for each request, and only
		// temporarily stores the dump content in memory, so we can call
		// resp.Dump() to get the dump content when needed in response
		// middleware.
		// This is actually a syntax sugar, implemented internally using
		// request middleware
		EnableDumpEachRequest().
		// Unmarshal response body into an APIError struct when status >= 400.
		SetCommonErrorResult(&APIError{}).
		// Handle common exceptions in response middleware.
		OnAfterResponse(func(client *req.Client, resp *req.Response) error {
			if resp.Err != nil { // There is an underlying error, e.g. network error or unmarshal error (SetSuccessResult or SetErrorResult was invoked before).
				if dump := resp.Dump(); dump != "" { // Append dump content to original underlying error to help troubleshoot.
					resp.Err = fmt.Errorf("error: %s\nraw content:\n%s", resp.Err.Error(), resp.Dump())
				}
				return nil // Skip the following logic if there is an underlying error.
			}
			if err, ok := resp.ErrorResult().(*APIError); ok { // Server returns an error message.
				// Convert it to human-readable go error which implements the error interface.
				resp.Err = err
				return nil
			}
			// Corner case: neither an error response nor a success response, e.g. status code < 200 or
			// code >= 300 && code <= 399, just dump the raw content into error to help troubleshoot.
			if !resp.IsSuccessState() {
				resp.Err = fmt.Errorf("bad response, raw content:\n%s", resp.Dump())
			}
			return nil
		})

	return &Client{
		Client: c,
	}
}

type apiNameType int

const apiNameKey apiNameType = iota

// SetTracer set the tracer of opentelemetry.
func (c *Client) SetTracer(tracer trace.Tracer) {
	c.WrapRoundTripFunc(func(rt req.RoundTripper) req.RoundTripFunc {
		return func(req *req.Request) (resp *req.Response, err error) {
			ctx := req.Context()
			apiName, ok := ctx.Value(apiNameKey).(string)
			if !ok {
				apiName = req.URL.Path
			}
			_, span := tracer.Start(req.Context(), apiName)
			defer span.End()
			span.SetAttributes(
				attribute.String("http.url", req.URL.String()),
				attribute.String("http.method", req.Method),
				attribute.String("http.req.header", req.HeaderToString()),
			)
			if len(req.Body) > 0 {
				span.SetAttributes(
					attribute.String("http.req.body", string(req.Body)),
				)
			}
			resp, err = rt.RoundTrip(req)
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
			}
			if resp.Response != nil {
				span.SetAttributes(
					attribute.Int("http.status_code", resp.StatusCode),
					attribute.String("http.resp.header", resp.HeaderToString()),
					attribute.String("http.resp.body", resp.String()),
				)
			}
			return
		}
	})
}

func withAPIName(ctx context.Context, name string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, apiNameKey, name)
}

type UserProfile struct {
	Name string `json:"name"`
	Blog string `json:"blog"`
}

// GetUserProfile returns the user profile for the specified user.
// Github API doc: https://docs.github.com/en/rest/users/users#get-a-user
func (c *Client) GetUserProfile(ctx context.Context, username string) (user *UserProfile, err error) {
	err = c.Get("/users/{username}").
		SetPathParam("username", username).
		SetSuccessResult(&user).
		Do(withAPIName(ctx, "GetUserProfile")).Err
	return
}

type Repo struct {
	Name string `json:"name"`
	Star int    `json:"stargazers_count"`
}

// ListUserRepo returns a list of public repositories for the specified user
// Github API doc: https://docs.github.com/en/rest/repos/repos#list-repositories-for-a-user
func (c *Client) ListUserRepo(ctx context.Context, username string, page int) (repos []*Repo, err error) {
	err = c.Get("/users/{username}/repos").
		SetPathParam("username", username).
		SetQueryParamsAnyType(map[string]any{
			"type":      "owner",
			"page":      strconv.Itoa(page),
			"per_page":  "100",
			"sort":      "updated",
			"direction": "desc",
		}).
		Do(withAPIName(ctx, "ListUserRepo")).
		Into(&repos)
	return
}

// LoginWithToken login with GitHub personal access token.
// GitHub API doc: https://docs.github.com/en/rest/overview/other-authentication-methods#authenticating-for-saml-sso
func (c *Client) LoginWithToken(token string) *Client {
	c.SetCommonHeader("Authorization", "token "+token)
	return c
}

// SetDebug enable debug if set to true, disable debug if set to false.
func (c *Client) SetDebug(enable bool) *Client {
	if enable {
		c.EnableDebugLog()
		c.EnableDumpAll()
	} else {
		c.DisableDebugLog()
		c.DisableDumpAll()
	}
	return c
}
