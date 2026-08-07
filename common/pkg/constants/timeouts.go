package constants

import "time"

const (
	HTTPTimeoutAuth     = 5 * time.Second
	HTTPTimeoutDefault  = 10 * time.Second
	HTTPTimeoutLong     = 30 * time.Second
	HTTPTimeoutPylovo   = 0 // no timeout; large models (750+ buildings) need unlimited time
	HTTPTimeoutExternal      = 15 * time.Second
	HTTPTimeoutComputeEngine = 10 * time.Second
)
