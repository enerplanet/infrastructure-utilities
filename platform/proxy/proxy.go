package proxy

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// HeaderMutator copies headers from src to dst with custom filtering logic.
type HeaderMutator func(dst, src http.Header)

// TargetBuilder constructs the upstream URL using the provided baseURL and the incoming request context.
type TargetBuilder func(baseURL string, c *gin.Context) string

// HandlerOptions customises the behaviour of the generated proxy handler.
type HandlerOptions struct {
	Timeout               time.Duration
	Component             string
	ForwardCookies        bool
	StripPrefix           string
	TargetBuilder         TargetBuilder
	RequestHeaderMutator  HeaderMutator
	ResponseHeaderMutator HeaderMutator
	Logger                logrus.FieldLogger
}

const defaultProxyTimeout = 30 * time.Second

// ErrInvalidBaseURL is returned when the provided base URL is empty.
var ErrInvalidBaseURL = fmt.Errorf("proxy: baseURL must not be empty")

// NewHandler builds a generic reverse proxy handler that forwards traffic to baseURL using
// the supplied options.
func NewHandler(baseURL string, opts HandlerOptions) (gin.HandlerFunc, error) {
	trimmedBase, err := normalizeBaseURL(baseURL)
	if err != nil {
		return nil, err
	}

	builder := builderOrDefault(opts.TargetBuilder, opts.StripPrefix)
	reqMutator := headerMutatorOrDefault(opts.RequestHeaderMutator)
	respMutator := headerMutatorOrDefault(opts.ResponseHeaderMutator)
	logger := loggerOrDefault(opts.Logger)
	component := componentOrDefault(opts.Component)
	client := &http.Client{Timeout: effectiveTimeout(opts.Timeout)}

	return func(c *gin.Context) {
		target := builder(trimmedBase, c)
		proxyReq, err := http.NewRequestWithContext(c.Request.Context(), c.Request.Method, target, c.Request.Body)
		if err != nil {
			logger.WithField("component", component).WithError(err).Error("failed to create proxy request")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create proxy request"})
			return
		}

		reqMutator(proxyReq.Header, c.Request.Header)
		if opts.ForwardCookies {
			for _, cookie := range c.Request.Cookies() {
				proxyReq.AddCookie(cookie)
			}
		}

		resp, err := client.Do(proxyReq)
		if err != nil {
			logger.WithField("component", component).WithError(err).Error("failed to reach upstream service")
			c.JSON(http.StatusBadGateway, gin.H{"error": "failed to reach upstream service"})
			return
		}
		defer func() { _ = resp.Body.Close() }()

		respMutator(c.Writer.Header(), resp.Header)
		c.Status(resp.StatusCode)
		if _, err := io.Copy(c.Writer, resp.Body); err != nil {
			logger.WithField("component", component).WithError(err).Warn("failed to stream upstream response")
		}
	}, nil
}

func normalizeBaseURL(baseURL string) (string, error) {
	trimmed := strings.TrimRight(baseURL, "/")
	if trimmed == "" {
		return "", ErrInvalidBaseURL
	}
	return trimmed, nil
}

func effectiveTimeout(spec time.Duration) time.Duration {
	if spec > 0 {
		return spec
	}
	return defaultProxyTimeout
}

func builderOrDefault(builder TargetBuilder, stripPrefix string) TargetBuilder {
	if builder != nil {
		return builder
	}
	return passthroughBuilder(stripPrefix)
}

func headerMutatorOrDefault(mut HeaderMutator) HeaderMutator {
	if mut != nil {
		return mut
	}
	return CopyAllHeaders
}

func loggerOrDefault(logger logrus.FieldLogger) logrus.FieldLogger {
	if logger != nil {
		return logger
	}
	return logrus.StandardLogger()
}

func componentOrDefault(component string) string {
	if component != "" {
		return component
	}
	return "proxy"
}

// MustNewHandler behaves like NewHandler but panics when the handler cannot be created.
func MustNewHandler(baseURL string, opts HandlerOptions) gin.HandlerFunc {
	handler, err := NewHandler(baseURL, opts)
	if err != nil {
		panic(err)
	}
	return handler
}

// CopyAllHeaders copies every header from src to dst without filtering.
func CopyAllHeaders(dst, src http.Header) {
	copyHeadersWithFilter(dst, src, nil)
}

// CopyHeadersExcept copies headers excluding the provided case-insensitive keys.
func CopyHeadersExcept(excluded ...string) HeaderMutator {
	skip := make(map[string]struct{}, len(excluded))
	for _, key := range excluded {
		skip[strings.ToLower(key)] = struct{}{}
	}
	return func(dst, src http.Header) {
		copyHeadersWithFilter(dst, src, func(lowerKey string) bool {
			_, ok := skip[lowerKey]
			return ok
		})
	}
}

// CopyHeadersSkipping drops headers whose lower-case names exist in the skip set.
func CopyHeadersSkipping(skip map[string]struct{}) HeaderMutator {
	if len(skip) == 0 {
		return CopyAllHeaders
	}

	normalized := make(map[string]struct{}, len(skip))
	for key := range skip {
		normalized[strings.ToLower(key)] = struct{}{}
	}

	return func(dst, src http.Header) {
		copyHeadersWithFilter(dst, src, func(lowerKey string) bool {
			_, ok := normalized[lowerKey]
			return ok
		})
	}
}

func copyHeadersWithFilter(dst, src http.Header, shouldSkip func(lowerKey string) bool) {
	for key, values := range src {
		lowerKey := strings.ToLower(key)
		if shouldSkip != nil && shouldSkip(lowerKey) {
			continue
		}
		for _, value := range values {
			dst.Add(key, value)
		}
	}
}

// passthroughBuilder appends the (optionally stripped) request path and query string to the baseURL.
func passthroughBuilder(stripPrefix string) TargetBuilder {
	prefix := strings.TrimSuffix(stripPrefix, "/")
	return func(baseURL string, c *gin.Context) string {
		path := c.Request.URL.Path
		if prefix != "" && strings.HasPrefix(path, prefix) {
			path = strings.TrimPrefix(path, prefix)
			if !strings.HasPrefix(path, "/") {
				path = "/" + path
			}
		}

		trimmedBase := strings.TrimRight(baseURL, "/")
		if path == "" {
			path = "/"
		}
		target := trimmedBase + path
		if raw := c.Request.URL.RawQuery; raw != "" {
			target += "?" + raw
		}
		return target
	}
}
