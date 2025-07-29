package types

import "fmt"

type ServiceName string

var SSHDService ServiceName = "sshd"

func (s ServiceName) String() string {
	return string(s)
}

func ParseServiceName(service string) (ServiceName, bool) {
	switch service {
	case SSHDService.String():
		return SSHDService, true
	default:
		return ServiceName(""), false
	}
}

func ParseServiceNameFromAny(value any) (ServiceName, bool) {
	if str, ok := value.(string); ok {
		return ParseServiceName(str)
	}
	return ParseServiceName(fmt.Sprintf("%v", value))
}
