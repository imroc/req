package req

import (
	"bytes"
	"errors"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/imroc/req/v3/internal/tests"
)

func TestRetryBackOff(t *testing.T) {
	testRetry(t, func(r *Request) {
		r.SetRetryBackoffInterval(10*time.Millisecond, 1*time.Second)
	})
}

func testRetry(t *testing.T, setFunc func(r *Request)) {
	attempt := 0
	r := tc().R().
		SetRetryCount(3).
		SetRetryCondition(func(resp *Response, err error) bool {
			return (err != nil) || (resp.StatusCode == http.StatusTooManyRequests)
		}).
		SetRetryHook(func(resp *Response, err error) {
			attempt++
		})
	setFunc(r)
	resp, err := r.Get("/too-many")
	tests.AssertNoError(t, err)
	tests.AssertEqual(t, 3, resp.Request.RetryAttempt)
	tests.AssertEqual(t, 3, attempt)
}

func TestRetryInterval(t *testing.T) {
	testRetry(t, func(r *Request) {
		r.SetRetryInterval(func(resp *Response, attempt int) time.Duration {
			sleep := 0.01 * math.Exp2(float64(attempt))
			return time.Duration(math.Min(2, sleep)) * time.Second
		})
	})
}

func TestRetryFixedInterval(t *testing.T) {
	testRetry(t, func(r *Request) {
		r.SetRetryFixedInterval(1 * time.Millisecond)
	})
}

func TestAddRetryHook(t *testing.T) {
	test := "test1"
	testRetry(t, func(r *Request) {
		r.AddRetryHook(func(resp *Response, err error) {
			test = "test2"
		})
	})
	tests.AssertEqual(t, "test2", test)
}

func TestRetryOverride(t *testing.T) {
	c := tc().
		SetCommonRetryCount(3).
		SetCommonRetryHook(func(resp *Response, err error) {}).
		AddCommonRetryHook(func(resp *Response, err error) {}).
		SetCommonRetryCondition(func(resp *Response, err error) bool {
			return false
		}).SetCommonRetryBackoffInterval(1*time.Millisecond, 10*time.Millisecond)
	test := "test"
	resp, err := c.R().SetRetryFixedInterval(2 * time.Millisecond).
		SetRetryCount(2).
		SetRetryHook(func(resp *Response, err error) {
			test = "test1"
		}).SetRetryCondition(func(resp *Response, err error) bool {
		return err != nil || resp.StatusCode == http.StatusTooManyRequests
	}).Get("/too-many")
	tests.AssertNoError(t, err)
	tests.AssertEqual(t, "test1", test)
	tests.AssertEqual(t, 2, resp.Request.RetryAttempt)
}

func TestAddRetryCondition(t *testing.T) {
	attempt := 0
	resp, err := tc().R().
		SetRetryCount(3).
		AddRetryCondition(func(resp *Response, err error) bool {
			return err != nil
		}).
		AddRetryCondition(func(resp *Response, err error) bool {
			return resp.StatusCode == http.StatusServiceUnavailable
		}).
		SetRetryHook(func(resp *Response, err error) {
			attempt++
		}).Get("/too-many")
	tests.AssertNoError(t, err)
	tests.AssertEqual(t, 0, attempt)
	tests.AssertEqual(t, 0, resp.Request.RetryAttempt)

	attempt = 0
	resp, err = tc().
		SetCommonRetryCount(3).
		AddCommonRetryCondition(func(resp *Response, err error) bool {
			return err != nil
		}).
		AddCommonRetryCondition(func(resp *Response, err error) bool {
			return resp.StatusCode == http.StatusServiceUnavailable
		}).
		SetCommonRetryHook(func(resp *Response, err error) {
			attempt++
		}).R().Get("/too-many")
	tests.AssertNoError(t, err)
	tests.AssertEqual(t, 0, attempt)
	tests.AssertEqual(t, 0, resp.Request.RetryAttempt)

}

func TestRetryWithUnreplayableBody(t *testing.T) {
	_, err := tc().R().
		SetRetryCount(1).
		SetBody(bytes.NewBufferString("test")).
		Post("/")
	tests.AssertEqual(t, errRetryableWithUnReplayableBody, err)

	_, err = tc().R().
		SetRetryCount(1).
		SetBody(io.NopCloser(bytes.NewBufferString("test"))).
		Post("/")
	tests.AssertEqual(t, errRetryableWithUnReplayableBody, err)
}

func TestRetryWithSetResult(t *testing.T) {
	headers := make(http.Header)
	resp, err := tc().SetCommonCookies(&http.Cookie{
		Name:  "test",
		Value: "test",
	}).R().
		SetRetryCount(1).
		SetResult(&headers).
		Get("/header")
	assertSuccess(t, resp, err)
	tests.AssertEqual(t, "test=test", headers.Get("Cookie"))
}

func TestRetryWithModify(t *testing.T) {
	tokens := []string{"badtoken1", "badtoken2", "goodtoken"}
	tokenIndex := 0
	c := tc().
		SetCommonRetryCount(2).
		SetCommonRetryHook(func(resp *Response, err error) {
			tokenIndex++
			resp.Request.SetBearerAuthToken(tokens[tokenIndex])
		}).SetCommonRetryCondition(func(resp *Response, err error) bool {
		return err != nil || resp.StatusCode == http.StatusUnauthorized
	})

	resp, err := c.R().
		SetBearerAuthToken(tokens[tokenIndex]).
		Get("/protected")
	assertSuccess(t, resp, err)
	tests.AssertEqual(t, 2, resp.Request.RetryAttempt)
}

func TestRetryOnBeforeRequestError(t *testing.T) {
	failCount := 0
	retryHookCount := 0
	c := tc().OnBeforeRequest(func(client *Client, request *Request) error {
		failCount++
		if failCount < 3 {
			return errors.New("temporary before-request error")
		}
		return nil
	})
	resp, err := c.R().
		SetRetryCount(2).
		SetRetryFixedInterval(1 * time.Millisecond).
		SetRetryCondition(func(resp *Response, err error) bool {
			return err != nil && err.Error() == "temporary before-request error"
		}).
		SetRetryHook(func(resp *Response, err error) {
			retryHookCount++
		}).
		Get("/")
	assertSuccess(t, resp, err)
	tests.AssertEqual(t, 2, resp.Request.RetryAttempt)
	tests.AssertEqual(t, 2, retryHookCount)
	tests.AssertEqual(t, 3, failCount)
}

func TestRetryConditionHasRequestOnBeforeRequestError(t *testing.T) {
	c := tc().OnBeforeRequest(func(client *Client, request *Request) error {
		return errors.New("before-request error")
	})
	resp, err := c.R().
		SetRetryCount(0).
		SetRetryCondition(func(resp *Response, err error) bool {
			tests.AssertNotNil(t, resp)
			tests.AssertNotNil(t, resp.Request)
			tests.AssertEqual(t, "/header", resp.Request.RawURL)
			tests.AssertEqual(t, http.MethodGet, resp.Request.Method)
			tests.AssertEqual(t, err, resp.Err)
			return false
		}).
		Get("/header")
	tests.AssertNotNil(t, err)
	tests.AssertNotNil(t, resp.Request)
}

func TestNoRetryOnBeforeRequestErrorWhenConditionFalse(t *testing.T) {
	c := tc().OnBeforeRequest(func(client *Client, request *Request) error {
		return errors.New("not retryable")
	})
	resp, err := c.R().
		SetRetryCount(3).
		SetRetryCondition(func(resp *Response, err error) bool {
			return false
		}).
		Get("/")
	tests.AssertNotNil(t, err)
	tests.AssertEqual(t, 0, resp.Request.RetryAttempt)
}

func TestRetryFalse(t *testing.T) {
	resp, err := tc().SetTimeout(2 * time.Second).R().
		SetRetryCount(1).
		SetRetryCondition(func(resp *Response, err error) bool {
			return false
		}).Get("https://non-exists-host.com.cn")
	tests.AssertNotNil(t, err)
	tests.AssertIsNil(t, resp.Response)
	tests.AssertEqual(t, 0, resp.Request.RetryAttempt)
}

func TestRetryTurnedOffWhenRetryCountEqZero(t *testing.T) {
	resp, err := tc().SetTimeout(2 * time.Second).R().
		SetRetryCount(0).
		SetRetryCondition(func(resp *Response, err error) bool {
			t.Fatal("retry condition should not be executed")
			return true
		}).
		Get("https://non-exists-host.com.cn")
	tests.AssertNotNil(t, err)
	tests.AssertIsNil(t, resp.Response)
	tests.AssertEqual(t, 0, resp.Request.RetryAttempt)

	resp, err = tc().SetTimeout(2 * time.Second).
		SetCommonRetryCount(0).
		SetCommonRetryCondition(func(resp *Response, err error) bool {
			t.Fatal("retry condition should not be executed")
			return true
		}).
		R().
		Get("https://non-exists-host.com.cn")
	tests.AssertNotNil(t, err)
	tests.AssertIsNil(t, resp.Response)
	tests.AssertEqual(t, 0, resp.Request.RetryAttempt)
}

func TestGetRetryOptionNilWhenNotConfigured(t *testing.T) {
	r := tc().R()
	tests.AssertIsNil(t, r.GetRetryOption())
}

func TestGetRetryOptionFromRequest(t *testing.T) {
	r := tc().R().SetRetryCount(5)
	ro := r.GetRetryOption()
	tests.AssertNotNil(t, ro)
	tests.AssertEqual(t, 5, ro.MaxRetries)
}

func TestGetRetryOptionFromClient(t *testing.T) {
	c := tc().SetCommonRetryCount(4)
	r := c.R()
	ro := r.GetRetryOption()
	tests.AssertNotNil(t, ro)
	tests.AssertEqual(t, 4, ro.MaxRetries)
}

func TestGetRetryOptionRequestOverridesClient(t *testing.T) {
	c := tc().SetCommonRetryCount(4)
	r := c.R().SetRetryCount(1)
	ro := r.GetRetryOption()
	tests.AssertNotNil(t, ro)
	tests.AssertEqual(t, 1, ro.MaxRetries)
	// Client-level option remains unchanged
	tests.AssertEqual(t, 4, c.getRetryOption().MaxRetries)
}

// TestGetRetryOptionInMiddleware covers the #475 use case: middleware reads
// MaxRetries so exception reporting only runs after all retries are exhausted.
func TestGetRetryOptionInMiddleware(t *testing.T) {
	reportCount := 0
	middlewareCalls := 0
	var seenMaxRetries int

	c := tc().
		SetCommonRetryCount(2).
		SetCommonRetryFixedInterval(1 * time.Millisecond).
		SetCommonRetryCondition(func(resp *Response, err error) bool {
			return err != nil || resp.StatusCode == http.StatusTooManyRequests
		}).
		OnAfterResponse(func(client *Client, resp *Response) error {
			middlewareCalls++
			ro := resp.Request.GetRetryOption()
			tests.AssertNotNil(t, ro)
			seenMaxRetries = ro.MaxRetries

			// Report only when retry budget is exhausted and the attempt failed
			// (HTTP error status or transport error).
			failed := resp.IsErrorState() || resp.Err != nil
			if failed &&
				resp.Request.RetryAttempt >= ro.MaxRetries &&
				ro.MaxRetries >= 0 {
				reportCount++
			}
			return nil
		})

	resp, err := c.R().Get("/too-many")

	tests.AssertNoError(t, err)
	tests.AssertEqual(t, 2, seenMaxRetries)
	tests.AssertEqual(t, 2, resp.Request.RetryAttempt)
	// Initial attempt + 2 retries => 3 middleware invocations
	tests.AssertEqual(t, 3, middlewareCalls)
	// Only the final exhausted attempt should report once
	tests.AssertEqual(t, 1, reportCount)
}

func TestGetRetryOptionInMiddlewareTransportError(t *testing.T) {
	// Transport failures leave resp.Response nil so IsErrorState is false;
	// reporting must also consider resp.Err.
	reportCount := 0
	middlewareCalls := 0
	c := C().
		SetTimeout(500 * time.Millisecond).
		SetCommonRetryCount(1).
		SetCommonRetryFixedInterval(1 * time.Millisecond).
		OnAfterResponse(func(client *Client, resp *Response) error {
			middlewareCalls++
			ro := resp.Request.GetRetryOption()
			tests.AssertNotNil(t, ro)
			failed := resp.IsErrorState() || resp.Err != nil
			if failed &&
				resp.Request.RetryAttempt >= ro.MaxRetries &&
				ro.MaxRetries >= 0 {
				reportCount++
			}
			return nil
		})

	resp, err := c.R().Get("https://non-exists-host.com.cn")
	tests.AssertNotNil(t, err)
	tests.AssertEqual(t, 1, resp.Request.RetryAttempt)
	tests.AssertEqual(t, 2, middlewareCalls) // initial + 1 retry
	tests.AssertEqual(t, 1, reportCount)
}

func TestGetRetryOptionInMiddlewareNoReportOnEventualSuccess(t *testing.T) {
	// Server fails twice then succeeds; middleware must not report after success.
	attempt := 0
	reportCount := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempt++
		if attempt <= 2 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer ts.Close()

	c := C().
		SetCommonRetryCount(3).
		SetCommonRetryFixedInterval(1 * time.Millisecond).
		SetCommonRetryCondition(func(resp *Response, err error) bool {
			return err != nil || resp.StatusCode == http.StatusServiceUnavailable
		})

	resp, err := c.R().
		OnAfterResponse(func(client *Client, resp *Response) error {
			ro := resp.Request.GetRetryOption()
			if ro == nil {
				return nil
			}
			failed := resp.IsErrorState() || resp.Err != nil
			if failed &&
				resp.Request.RetryAttempt >= ro.MaxRetries &&
				ro.MaxRetries >= 0 {
				reportCount++
			}
			return nil
		}).
		Get(ts.URL)

	tests.AssertNoError(t, err)
	tests.AssertEqual(t, http.StatusOK, resp.StatusCode)
	tests.AssertEqual(t, 0, reportCount)
	tests.AssertEqual(t, 2, resp.Request.RetryAttempt)
}
