package client

import "time"

const (
	HTTPClientTimeout  = time.Millisecond * 500000
	HTTPRequestTimeout = time.Millisecond * 3000
)
