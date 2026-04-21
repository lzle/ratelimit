package ratelimit

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestNewRateLimiter(t *testing.T) {
	limiter, err := NewRateLimiter(nil)
	if err == nil {
		t.Fatalf("expected error for empty configs")
	}
	if limiter != nil {
		t.Fatalf("expected nil limiter when configs are empty")
	}

	limiter, err = NewRateLimiter(map[string]RateLimitConfig{
		"user-a": {BandwidthQuota: 1024},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if limiter == nil {
		t.Fatalf("expected non-nil limiter")
	}
	if limiter.keyFlowCtrl == nil {
		t.Fatalf("expected initialized keyFlowCtrl")
	}
}

func TestGetReader(t *testing.T) {
	limiter, err := NewRateLimiter(map[string]RateLimitConfig{
		"user-a": {BandwidthQuota: 256},
		"user-b": {BandwidthQuota: 0},
	})
	if err != nil {
		t.Fatalf("new limiter failed: %v", err)
	}

	rawReader := strings.NewReader("hello")
	got := limiter.GetReader("missing", rawReader)
	if got != rawReader {
		t.Fatalf("expected original reader for missing key")
	}

	rawReaderZero := strings.NewReader("world")
	got = limiter.GetReader("user-b", rawReaderZero)
	if got != rawReaderZero {
		t.Fatalf("expected original reader for zero quota")
	}

	rawReaderLimited := strings.NewReader("ratelimit")
	got = limiter.GetReader("user-a", rawReaderLimited)
	if got == rawReaderLimited {
		t.Fatalf("expected wrapped reader for non-zero quota")
	}

	data, readErr := io.ReadAll(got)
	if readErr != nil {
		t.Fatalf("read limited reader failed: %v", readErr)
	}
	if string(data) != "ratelimit" {
		t.Fatalf("unexpected data: %q", string(data))
	}
}

func TestGetWriter(t *testing.T) {
	limiter, err := NewRateLimiter(map[string]RateLimitConfig{
		"user-a": {BandwidthQuota: 256},
		"user-b": {BandwidthQuota: 0},
	})
	if err != nil {
		t.Fatalf("new limiter failed: %v", err)
	}

	rawWriter := &bytes.Buffer{}
	got := limiter.GetWriter("missing", rawWriter)
	if got != rawWriter {
		t.Fatalf("expected original writer for missing key")
	}

	rawWriterZero := &bytes.Buffer{}
	got = limiter.GetWriter("user-b", rawWriterZero)
	if got != rawWriterZero {
		t.Fatalf("expected original writer for zero quota")
	}

	rawWriterLimited := &bytes.Buffer{}
	got = limiter.GetWriter("user-a", rawWriterLimited)
	if got == rawWriterLimited {
		t.Fatalf("expected wrapped writer for non-zero quota")
	}

	_, writeErr := got.Write([]byte("ratelimit"))
	if writeErr != nil {
		t.Fatalf("write limited writer failed: %v", writeErr)
	}
	if rawWriterLimited.String() != "ratelimit" {
		t.Fatalf("unexpected write result: %q", rawWriterLimited.String())
	}
}

func TestClose(t *testing.T) {
	limiter, err := NewRateLimiter(map[string]RateLimitConfig{
		"user-a": {BandwidthQuota: 256},
	})
	if err != nil {
		t.Fatalf("new limiter failed: %v", err)
	}

	_ = limiter.GetReader("user-a", strings.NewReader("hello"))
	if err := limiter.Close("user-a"); err != nil {
		t.Fatalf("expected first close success, got %v", err)
	}
	if err := limiter.Close("user-a"); err == nil {
		t.Fatalf("expected second close to return error")
	}
	if err := limiter.Close("missing"); err == nil {
		t.Fatalf("expected missing key close to return error")
	}
}
