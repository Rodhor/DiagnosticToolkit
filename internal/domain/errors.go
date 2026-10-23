package domain

import "errors"

var ErrDNSHijack = errors.New("lookup succeeded for non-existent name (possible DNS hijack)")
