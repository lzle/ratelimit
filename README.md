# ratelimit

[![Go Reference](https://pkg.go.dev/badge/github.com/lzle/ratelimit.svg)](https://pkg.go.dev/github.com/lzle/ratelimit)
[![Go](https://img.shields.io/badge/go-1.22+-00ADD8?logo=go)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

**Token-bucket** rate limiting for Go `io.Reader` / `io.Writer`, with **per-key** isolation and thread-safe controllers.

## Highlights

| | |
|---|---|
| **Per-key limits** | Same key shares one limiter; different keys are independent |
| **Stream-first** | Wrap readers/writers; no protocol coupling |
| **Release** | Call `Close(key)` after use to avoid goroutine leaks |

## Install

```bash
go get github.com/lzle/ratelimit@latest
```

*Module:* `github.com/lzle/ratelimit` — see [`go.mod`](go.mod).

## Quick start

```go
limiter, err := ratelimit.NewRateLimiter(map[string]ratelimit.RateLimitConfig{
	"user-a": {BandwidthQuota: 1024}, // bytes per second
})
if err != nil {
	log.Fatal(err)
}

key := "user-a"
r := limiter.GetReader(key, strings.NewReader("hello"))
w := limiter.GetWriter(key, io.Discard)
_, _ = io.Copy(w, r)
_ = limiter.Close(key)
```

## API

| Method | Role |
|--------|------|
| `NewRateLimiter(configs)` | Build limiter; fails if `configs` is empty |
| `GetReader(key, r)` / `GetWriter(key, w)` | Limited I/O, or original if key missing / quota `0` |
| `Close(key)` | Release the controller for `key` (pair with each `Get*` that acquired) |

Lower-level helpers: package **`flowctrl`** (`NewRateReader`, `NewKeyFlowCtrl`, …).

## Notes

- **Quota** = **bytes per second** (`BandwidthQuota`).
- Each `GetReader` / `GetWriter` **acquires** the key — balance with **`Close(key)`** or references accumulate.
- Internals: token bucket, **50ms** refill window, **min burst 64** bytes per window.

## Test

```bash
go test ./... -count=1
```

## Acknowledgments

The [`flowctrl`](flowctrl/) package is largely based on [CubeFS](https://github.com/cubefs/cubefs) [`util/flowctrl`](https://github.com/cubefs/cubefs/tree/master/util/flowctrl). See [`flowctrl/NOTICE`](flowctrl/NOTICE) for attribution and upstream licensing.

## License

This repository is [MIT](LICENSE) licensed for the parts authored here (Copyright (c) 2026 lzle). Third-party code in `flowctrl/` remains subject to the [Apache License 2.0](https://github.com/cubefs/cubefs/blob/master/LICENSE) of the CubeFS project; see [`flowctrl/NOTICE`](flowctrl/NOTICE).
