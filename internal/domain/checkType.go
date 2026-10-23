package domain

type CheckType string

const (
	TCP  CheckType = "TCP"
	Ping CheckType = "Ping"
	DNS  CheckType = "DNS"
)
