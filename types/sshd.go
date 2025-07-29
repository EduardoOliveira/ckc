package types

type SSHDParsedEvent struct {
	Result    string `json:"result"`
	Success   bool   `json:"success"`
	Method    string `json:"method"`
	Signature string `json:"signature"`
}
