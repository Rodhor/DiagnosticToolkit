package domain

const (
	KeyReason            = "reason"
	KeyRemoteAddr        = "remote_addr"
	KeyWarnThreshold     = "warn_threshold"
	KeyResolver          = "resolver"
	KeyQueryName         = "query_name"
	KeyExpectedAddresses = "expected_addresses"
	KeyActualAddresses   = "actual_addresses"
)

const (
	ReasonTimeout          = "timeout"
	ReasonRefused          = "refused"
	ReasonUnknown          = "unknown"
	ReasonHighLatency      = "high_latency"
	ReasonDNSFailure       = "dns_failure"
	ReasonNXDomain         = "nx_domain"
	ReasonServerFailure    = "server_failure"
	ReasonUnexpectedAnswer = "unexpected_answer"
	ReasonNXDomainHijacked = "nx_domain_hijacked"
	ReasonReset            = "reset"             
	ReasonUnreachable      = "unreachable"       
	ReasonPermissionDenied = "permission_denied" 
	ReasonCanceled         = "canceled"
)
