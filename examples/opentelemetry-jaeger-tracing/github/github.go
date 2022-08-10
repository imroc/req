package github

import (
	"context"
	"fmt"
	"github.com/imroc/req/v3"
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
		// EnableDump at the request level in request middleware which dump content into
		// memory (not print to stdout), we can record dump content only when unexpected
		// exception occurs, it is helpful to troubleshoot problems in production.
		OnBeforeRequest(func(c *req.Client, r *req.Request) error {
			if r.RetryAttempt > 0 { // Ignore on retry.
				return nil
			}
			r.EnableDump()
			return nil
		}).
		// Unmarshal response body into an APIError struct when status >= 400.
		SetCommonError(&APIError{}).
		// Handle common exceptions in response middleware.
		OnAfterResponse(func(client *req.Client, resp *req.Response) error {
			if resp.Err != nil {
				if dump := resp.Dump(); dump != "" { // Append dump content to original underlying error to help troubleshoot.
					resp.Err = fmt.Errorf("%s\nraw dump:\n%s", resp.Err.Error(), resp.Dump())
				}
				return nil // Skip the following logic if there is an underlying error.
			}
			if err, ok := resp.Error().(*APIError); ok { // Server returns an error message.
				// Convert it to human-readable go error.
				resp.Err = err
				return nil
			}
			// Corner case: neither an error response nor a success response,
			// dump content to help troubleshoot.
			if !resp.IsSuccess() {
				resp.Err = fmt.Errorf("bad response, raw dump:\n%s", resp.Dump())
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
		return func(r *req.Request) (resp *req.Response, err error) {
			ctx := r.Context()
			spanName := ctx.Value(apiNameKey).(string)
			_, span := tracer.Start(r.Context(), spanName)
			defer span.End()
			span.SetAttributes(
				attribute.String("http.url", r.URL.String()),
				attribute.String("http.method", r.Method),
				attribute.String("http.req.header", r.HeaderToString()),
			)
			if len(r.Body) > 0 {
				span.SetAttributes(
					attribute.String("http.req.body", string(r.Body)),
				)
			}
			resp, err = rt.RoundTrip(r)
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
				return
			}
			span.SetAttributes(
				attribute.Int("http.status_code", resp.StatusCode),
				attribute.String("http.resp.header", resp.HeaderToString()),
				attribute.String("resp.resp.body", resp.String()),
			)
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
		Do(withAPIName(ctx, "GetUserProfile")).
		Into(&user)
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
		SetQueryParamsAnyType(map[string]any{
			"type":      "owner",
			"page":      strconv.Itoa(page),
			"per_page":  "100",
			"sort":      "updated",
			"direction": "desc",
		}).
		SetPathParam("username", username).
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
