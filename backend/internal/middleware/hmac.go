// Package middleware provides reusable Gin middleware for the kube-rca backend.
//
// The HMAC middleware authenticates inbound webhook requests by comparing a
// hex-encoded HMAC-SHA256 signature carried in a configurable header against
// a digest computed over the raw request body using a shared secret.
//
// Opt-in semantics: when the configured secret is empty, the middleware logs
// a single startup warning (handled by the caller) and lets requests through
// unchanged. This preserves backward compatibility for operators that have
// not yet provisioned a webhook secret. When the secret is configured the
// middleware rejects requests with a missing or invalid signature using
// HTTP 401.
package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// DefaultWebhookSignatureHeader is the canonical header name used to carry the
// HMAC-SHA256 hex digest produced by the upstream sender.
const DefaultWebhookSignatureHeader = "X-Webhook-Signature"

// HMACConfig configures the HMACMiddleware.
//
// Secret: shared secret used to compute HMAC-SHA256(body). When empty the
// middleware operates in opt-in mode and forwards requests unchanged.
// HeaderName: optional override for the signature header. Defaults to
// DefaultWebhookSignatureHeader when empty.
type HMACConfig struct {
	Secret     string
	HeaderName string
}

// HMACMiddleware returns a Gin middleware that validates HMAC-SHA256
// signatures on the request body.
//
// The middleware drains and replaces c.Request.Body so downstream handlers
// can still call c.ShouldBindJSON or read the body normally. Signature
// comparison uses crypto/hmac.Equal to avoid timing leaks.
func HMACMiddleware(cfg HMACConfig) gin.HandlerFunc {
	header := cfg.HeaderName
	if header == "" {
		header = DefaultWebhookSignatureHeader
	}
	secret := []byte(cfg.Secret)

	return func(c *gin.Context) {
		// Always read & restore the body so handlers can re-read it.
		body, err := readAndRestoreBody(c.Request)
		if err != nil {
			log.Printf("hmac middleware: failed to read body: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			c.Abort()
			return
		}

		// Opt-in mode: when no secret is configured, allow the request.
		if len(secret) == 0 {
			c.Next()
			return
		}

		provided := strings.TrimSpace(c.GetHeader(header))
		// Strip optional "sha256=" prefix used by some senders (e.g. GitHub style).
		if eq := strings.Index(provided, "="); eq != -1 && strings.EqualFold(provided[:eq], "sha256") {
			provided = provided[eq+1:]
		}
		if provided == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing signature"})
			c.Abort()
			return
		}

		providedBytes, err := hex.DecodeString(provided)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid signature encoding"})
			c.Abort()
			return
		}

		mac := hmac.New(sha256.New, secret)
		mac.Write(body)
		expected := mac.Sum(nil)

		if !hmac.Equal(providedBytes, expected) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid signature"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// readAndRestoreBody reads the entire request body and replaces it so that
// downstream handlers (e.g. c.ShouldBindJSON) can read it again.
func readAndRestoreBody(req *http.Request) ([]byte, error) {
	if req.Body == nil {
		return nil, nil
	}
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	if cerr := req.Body.Close(); cerr != nil {
		log.Printf("hmac middleware: body close error: %v", cerr)
	}
	req.Body = io.NopCloser(bytes.NewReader(body))
	return body, nil
}
