package retryhttp

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/sabariramc/go-kit/log"
)

type Client struct {
	*http.Client
	log          *log.Logger
	retryMax     uint
	minRetryWait time.Duration
	maxRetryWait time.Duration
	checkRetry   CheckRetry
	backoff      Backoff
}

func New(options ...Option) *Client {
	config := GetDefaultConfig()
	for _, opt := range options {
		opt(&config)
	}
	return &Client{
		Client:       config.Client,
		retryMax:     config.RetryMax,
		minRetryWait: config.MinRetryWait,
		maxRetryWait: config.MaxRetryWait,
		checkRetry:   config.CheckRetry,
		backoff:      config.Backoff,
		log:          config.Log,
	}
}
func (c *Client) Get(ctx context.Context, url string) (resp *http.Response, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

func (c *Client) Post(ctx context.Context, url, contentType string, body io.Reader) (resp *http.Response, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	return c.Do(req)
}

func (c *Client) PostForm(ctx context.Context, url string, data url.Values) (resp *http.Response, err error) {
	return c.Post(ctx, url, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
}

func (c *Client) Head(ctx context.Context, url string) (resp *http.Response, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

// Do sends an HTTP request and performs retries with exponential backoff as needed,
// based on the retry and backoff configuration.
func (h *Client) Do(req *http.Request) (*http.Response, error) {
	/*this is a modified version of go-retryablehttp*/
	var resp *http.Response
	var attempt int
	var shouldRetry bool
	var doErr, respErr error
	var reqBody []byte
	if req.ContentLength > 0 {
		reqBody, _ = io.ReadAll(req.Body)
	}
	for i := 0; ; i++ {
		doErr = nil
		attempt++
		if req.ContentLength > 0 {
			req.Body = io.NopCloser(bytes.NewReader(reqBody))
		}
		resp, doErr = h.Client.Do(req)
		shouldRetry, respErr = h.backOffAndRetry(i, req, resp, doErr)
		if !shouldRetry {
			break
		}
	}

	// this is the closest we have to success criteria
	if doErr == nil && respErr == nil && !shouldRetry {
		return resp, nil
	}

	var err error
	if respErr != nil {
		err = respErr
	} else {
		err = doErr
	}

	return resp, err
}

// backOffAndRetry determines if the request should be retried and calculates the backoff duration.
// It logs the retry attempt and waits for the backoff duration before retrying.
func (h *Client) backOffAndRetry(i int, req *http.Request, resp *http.Response, doErr error) (bool, error) {
	shouldRetry, respErr := h.checkRetry(req.Context(), resp, doErr)
	if !shouldRetry || respErr != nil {
		return shouldRetry, respErr
	}
	remain := h.retryMax - uint(i)
	if remain <= 0 {
		return false, respErr
	}
	wait := h.backoff(h.minRetryWait, h.maxRetryWait, i, resp)
	if resp != nil && resp.ContentLength > 0 {
		defer resp.Body.Close()
		resBlob, _ := io.ReadAll(resp.Body)
		h.log.Warn(req.Context()).Str("response", string(resBlob)).Msgf("request failed with status code %v retry %v of %v in %vms, resp: ", resp.StatusCode, i+1, h.retryMax, wait.Milliseconds())
	} else if doErr != nil {
		h.log.Warn(req.Context()).Err(doErr).Msgf("request failed with error - retry %v of %v in %vms", i+1, h.retryMax, wait.Milliseconds())
	} else {
		h.log.Warn(req.Context()).Msgf("request failed - retry %v of %v in %vms", i+1, h.retryMax, wait.Milliseconds())
	}
	timer := time.NewTimer(wait)
	select {
	case <-req.Context().Done():
		timer.Stop()
		h.Client.CloseIdleConnections()
		return false, req.Context().Err()
	case <-timer.C:
	}
	return true, nil
}
