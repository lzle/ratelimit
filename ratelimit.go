package ratelimit

import (
	"fmt"
	"io"
	"strconv"

	"github/lzle/ratelimit/flowctrl"
)

type RateLimitConfig struct {
	BandwidthQuota uint64
}

type RateLimiter struct {
	keyFlowCtrl *flowctrl.KeyFlowCtrl
	configs     map[string]RateLimitConfig
}

func NewRateLimiter(configs map[string]RateLimitConfig) (*RateLimiter, error) {
	if len(configs) == 0 {
		return nil, fmt.Errorf("rate limit configs is empty")
	}

	return &RateLimiter{
		keyFlowCtrl: flowctrl.NewKeyFlowCtrl(),
		configs:     configs,
	}, nil
}

func (r *RateLimiter) GetReader(key string, reader io.Reader) io.Reader {
	cfg, ok := r.configs[key]
	if !ok {
		return reader
	}

	quota := cfg.BandwidthQuota
	if quota == 0 {
		return reader
	}

	rate, _ := convertUint64ToInt(quota)
	flowCtrl := r.keyFlowCtrl.Acquire(key, rate)
	reader = flowctrl.NewRateReaderWithCtrl(reader, flowCtrl)

	return reader
}

func (r *RateLimiter) GetWriter(key string, writer io.Writer) io.Writer {
	cfg, ok := r.configs[key]
	if !ok {
		return writer
	}
	quota := cfg.BandwidthQuota
	if quota == 0 {
		return writer
	}

	rate, _ := convertUint64ToInt(quota)
	flowCtrl := r.keyFlowCtrl.Acquire(key, rate)
	writer = flowctrl.NewRateWriterWithCtrl(writer, flowCtrl)

	return writer
}

func (r *RateLimiter) Close(key string) error {
	return r.keyFlowCtrl.Release(key)
}

func convertUint64ToInt(num uint64) (int, error) {
	str := strconv.FormatUint(num, 10)
	parsed, err := strconv.ParseInt(str, 10, 0)
	if err != nil {
		return 0, err
	}
	return int(parsed), nil
}
