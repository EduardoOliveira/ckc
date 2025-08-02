package types

import (
	"errors"
	"time"

	"github.com/EduardoOliveira/ckc/internal/maps"
	"github.com/EduardoOliveira/ckc/internal/opt"
)

var Now = time.Now

type ParsedEvent struct {
	ServiceName ServiceName `json:"service_name"`
	Hostname    string      `json:"hostname"`
	Ingestion   time.Time   `json:"ingestion"`
	IPAddress   IPAddress   `json:"ip_address"`
	Username    Username    `json:"username"`
	Service     Service     `json:"service"`

	// optional fields based on the event type
	SSHDEvent opt.Optional[SSHDParsedEvent] `json:"sshd_event"`
}

type IPAddress struct {
	Address    string               `json:"address"`
	Seen       int64                `json:"seen"`
	FirstSeen  time.Time            `json:"first_seen"`
	LastSeen   time.Time            `json:"last_seen"`
	EnrichedBy map[string]time.Time `json:"enriched_by,omitempty"`
}

func MapIPAddressFromMap(m map[string]any) (IPAddress, error) {
	var hasProps bool
	var found bool
	ip := IPAddress{}
	ip.Address, found = maps.GetValueAsString(m, "address")
	hasProps = hasProps || found
	ip.Seen, found = maps.GetValueAsInt64(m, "seen")
	hasProps = hasProps || found
	ip.FirstSeen, found = maps.GetValueAsTime(m, "first_seen")
	hasProps = hasProps || found
	ip.LastSeen, found = maps.GetValueAsTime(m, "last_seen")
	hasProps = hasProps || found
	if !hasProps {
		return IPAddress{}, errors.New("no properties found in map for IPAddress")
	}
	return ip, nil
}

type Country struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

type Service struct {
	Name string `json:"name"`
	Host string `json:"host,omitempty"`
	Port int    `json:"port"`
}

type WithUsername struct {
	IPAddress IPAddress `json:"ip_address"`
	Username  Username  `json:"username"`
	FirstTime time.Time `json:"first_time"`
	LastTime  time.Time `json:"last_time"`
}

type ConnectedTo struct {
	IPAddress IPAddress `json:"ip_address"`
	Service   Service   `json:"service"`
	FirstTime time.Time `json:"first_time"`
	LastTime  time.Time `json:"last_time"`
}

type Username struct {
	Name     string    `json:"name"`
	Seen     int64     `json:"seen"`
	FistSeen time.Time `json:"first_seen"`
	LastSeen time.Time `json:"last_seen"`
}

func MapUsernameFromMap(m map[string]any) (Username, error) {
	var hasProps bool
	var found bool
	u := Username{}
	u.Name, found = maps.GetValueAsString(m, "name")
	hasProps = hasProps || found
	u.Seen, found = maps.GetValueAsInt64(m, "seen")
	hasProps = hasProps || found
	u.FistSeen, found = maps.GetValueAsTime(m, "first_seen")
	hasProps = hasProps || found
	u.LastSeen, found = maps.GetValueAsTime(m, "last_seen")
	hasProps = hasProps || found
	if !hasProps {
		return Username{}, errors.New("no properties found in map for Username")
	}
	return u, nil
}

type AuthenticatedOn struct {
	Username   Username  `json:"username"`
	Service    Service   `json:"service"`
	FirstTime  time.Time `json:"first_time"`
	LastTime   time.Time `json:"last_time"`
	Failures   int       `json:"failures"`
	Successes  int       `json:"successes"`
	successful bool
}

func NewAuthenticatedOn(username Username, service Service, successful bool) AuthenticatedOn {
	return AuthenticatedOn{
		Username:   username,
		Service:    service,
		FirstTime:  Now(),
		LastTime:   Now(),
		successful: successful,
	}
}
