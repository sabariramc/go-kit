package retryhttp

import (
	"context"
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/sabariramc/go-kit/log"
)

// CheckRetry defines a function type for determining if a request should be retried.
type CheckRetry func(ctx context.Context, resp *http.Response, err error) (bool, error)

// Backoff defines a function type for determining the backoff duration between retries.
type Backoff func(min, max time.Duration, attemptNum int, resp *http.Response) time.Duration

// Config contains the configuration settings for the HTTP client, including logging,
// retry policies, backoff strategies, and the HTTP client itself.
type Config struct {
	RetryMax     uint          // RetryMax is the maximum number of retry attempts for failed requests.
	MinRetryWait time.Duration // MinRetryWait is the minimum duration to wait before retrying a failed request.
	MaxRetryWait time.Duration // MaxRetryWait is the maximum duration to wait before retrying a failed request.
	CheckRetry   CheckRetry    // CheckRetry is the function to determine if a request should be retried.
	Backoff      Backoff       // Backoff is the function to determine the wait duration between retries.
	Client       *http.Client  // Client is the underlying HTTP client used to make requests.
	Log          log.Logger    // Logger for the HTTP client
}

// newDefaultHTTPClient creates and configures a new HTTP client with custom transport settings.
func newDefaultHTTPClient() *http.Client {
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.MaxIdleConns = 100                 // MaxIdleConns sets the maximum number of idle connections across all hosts.
	t.MaxConnsPerHost = 100              // MaxConnsPerHost sets the maximum number of connections per host.
	t.MaxIdleConnsPerHost = 100          // MaxIdleConnsPerHost sets the maximum number of idle connections per host.
	t.IdleConnTimeout = 20 * time.Second // IdleConnTimeout sets the maximum amount of time an idle connection will remain open.
	return &http.Client{Transport: t,
		Timeout: 10 * time.Second, // Timeout sets the maximum duration for requests made by this client.
	}
}

// GetDefaultConfig returns a Config instance with default settings for the HTTP client.
func GetDefaultConfig() Config {
	return Config{
		RetryMax:     4,                                // Sets the maximum number of retry attempts to 4.
		MinRetryWait: time.Millisecond * 10,            // Sets the minimum retry wait duration to 10 milliseconds.
		MaxRetryWait: time.Second * 5,                  // Sets the maximum retry wait duration to 5 seconds.
		CheckRetry:   retryablehttp.DefaultRetryPolicy, // Uses the default retry policy.
		Backoff:      retryablehttp.DefaultBackoff,     // Uses the default backoff strategy.
		Client:       newDefaultHTTPClient(),           // Uses a custom HTTP client with specific transport settings.
		Log:          log.New("HttpClient"),
	}
}

// Option represents an option function for configuring the config struct.
type Option func(*Config)

// WithRetryMax sets the maximum number of retries for the config.
func WithRetryMax(retryMax uint) Option {
	return func(c *Config) {
		c.RetryMax = retryMax
	}
}

// WithMinRetryWait sets the minimum retry wait duration for the config.
func WithMinRetryWait(minRetryWait time.Duration) Option {
	return func(c *Config) {
		c.MinRetryWait = minRetryWait
	}
}

// WithMaxRetryWait sets the maximum retry wait duration for the config.
func WithMaxRetryWait(maxRetryWait time.Duration) Option {
	return func(c *Config) {
		c.MaxRetryWait = maxRetryWait
	}
}

// WithCheckRetry sets the check retry function for the config.
func WithCheckRetry(checkRetry CheckRetry) Option {
	return func(c *Config) {
		c.CheckRetry = checkRetry
	}
}

// WithBackoff sets the backoff strategy for the config.
func WithBackoff(backoff Backoff) Option {
	return func(c *Config) {
		c.Backoff = backoff
	}
}

// WithHTTPClient sets the HTTP client for the config.
func WithHTTPClient(client *http.Client) Option {
	return func(c *Config) {
		c.Client = client
	}
}
