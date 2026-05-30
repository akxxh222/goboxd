package config

import "time"

const (
	BuildTimeout = 10 * time.Second
	RunTimeout   = 3 * time.Second

	StaleTempDirAge = 30 * time.Minute

	MaxRequestBodyBytes  = 1 << 20
	MaxSourceBytes       = 256 << 10
	MaxTests             = 25
	MaxTestInputBytes    = 64 << 10
	MaxExpectedBytes     = 64 << 10
	MaxCapturedOutputLen = 64 << 10

	OutputTruncatedMarker = "\n[output truncated]\n"
)
