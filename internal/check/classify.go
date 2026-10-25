// internal/check/classify.go
package check

import (
	"context"
	"errors"
	"net"
	"os"
	"strings"
	"syscall"

	"github.com/rodhor/DiagnosticToolset/internal/domain"
)

// ClassifyError inspects a network error and returns a normalized reason token.
//
// ORDER-SENSITIVE: Specific error types must precede generic interfaces.
// For example, *net.DNSError implements net.Error and reports Timeout() == true.
// Testing net.Error first would misclassify DNS timeouts as firewall/packet drop timeouts.
func ClassifyError(err error) string {
	if err == nil {
		return ""
	}

	// 1. Process / context canceled
	if errors.Is(err, context.Canceled) {
		return domain.ReasonCanceled
	}

	// 2. Transport & Socket Level Errors (UNIX & Windows)
	if errors.Is(err, syscall.ECONNREFUSED) {
		return domain.ReasonRefused
	}

	if errors.Is(err, syscall.ECONNRESET) {
		return domain.ReasonReset
	}

	if errors.Is(err, syscall.EHOSTUNREACH) || errors.Is(err, syscall.ENETUNREACH) {
		return domain.ReasonUnreachable
	}

	if errors.Is(err, syscall.EACCES) || errors.Is(err, syscall.EPERM) {
		return domain.ReasonPermissionDenied
	}

	// 3. DNS Resolution Failures (must precede generic net.Error)
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		if dnsErr.IsNotFound {
			return domain.ReasonNXDomain
		}
		if dnsErr.IsTimeout {
			return domain.ReasonTimeout
		}
		// Inspect internal error string for transport failures wrapped inside DNSError
		errStr := strings.ToLower(dnsErr.Err)
		if strings.Contains(errStr, "connection refused") {
			return domain.ReasonRefused
		}
		if strings.Contains(errStr, "connection reset") {
			return domain.ReasonReset
		}
		if dnsErr.IsTemporary {
			return domain.ReasonServerFailure
		}
		return domain.ReasonDNSFailure
	}

	// 4. Generic Network Timeout (packet drop, firewall filter, host unreachable)
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return domain.ReasonTimeout
	}

	// 5. Fallback for unclassified errors
	return domain.ReasonUnknown
}

func isPermissionError(err error) bool {
	if errors.Is(err, os.ErrPermission) ||
		errors.Is(err, syscall.EPERM) ||
		errors.Is(err, syscall.EACCES) {
		return true
	}
	// Fallback check if persmission error
	return strings.Contains(strings.ToLower(err.Error()), "permission denied")
}
