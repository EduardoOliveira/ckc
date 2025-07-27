package types

import "time"

var Now = time.Now

type IPAddress struct {
	Address   string    `json:"address"`
	Seen      int       `json:"seen"`
	FirstSeen time.Time `json:"first_seen"`
	LastSeen  time.Time `json:"last_seen"`
}

func NewIPAddress(address string) IPAddress {
	return IPAddress{
		Address:   address,
		Seen:      1,
		FirstSeen: Now(),
		LastSeen:  Now(),
	}
}

func (ip *IPAddress) ToCypher() (cypher string, params map[string]any) {
	cypher = `
		MERGE (ip:IPAddress {address: $address})
		ON CREATE SET ip.seen = 0, ip.first_seen = datetime($ip_first_seen)
		SET ip.last_seen = datetime($ip_last_seen)
		SET ip.seen = ip.seen + 1 
		WITH ip
	`
	params = map[string]any{
		"address":       ip.Address,
		"ip_first_seen": ip.FirstSeen.Format(time.RFC3339),
		"ip_last_seen":  ip.LastSeen.Format(time.RFC3339),
	}
	return
}

type Service struct {
	Name string `json:"name"`
	Port int    `json:"port"`
}

func NewService(name string, port int) Service {
	return Service{
		Name: name,
		Port: port,
	}
}

func (s *Service) ToCypher() (cypher string, params map[string]any) {
	cypher = `
		MERGE (service:Service {name: $name, port: $port})
		WITH service
	`
	params = map[string]any{
		"name": s.Name,
		"port": s.Port,
	}
	return
}

type WithUsername struct {
	IPAddress IPAddress `json:"ip_address"`
	Username  Username  `json:"username"`
	FirstTime time.Time `json:"first_time"`
	LastTime  time.Time `json:"last_time"`
}

func NewWithUsername(ip IPAddress, username Username) WithUsername {
	return WithUsername{
		IPAddress: ip,
		Username:  username,
		FirstTime: Now(),
		LastTime:  Now(),
	}
}

func (w *WithUsername) ToCypher() (cypher string, params map[string]any) {
	cypher = `
		MATCH (ip:IPAddress {address: $wu_ip_address})
		MATCH (username:Username {name: $wu_username_name})
		MERGE (ip)-[w:WITH_USERNAME]->(username)
		ON CREATE SET w.first_time = datetime($wu_first_time), w.times = 0
		SET w.last_time = datetime($wu_last_time), w.times = w.times + 1
		WITH w	
	`
	params = map[string]any{
		"wu_ip_address":    w.IPAddress.Address,
		"wu_username_name": w.Username.Name,
		"wu_first_time":    w.FirstTime.Format(time.RFC3339),
		"wu_last_time":     w.LastTime.Format(time.RFC3339),
	}
	return
}

type ConnectedTo struct {
	IPAddress IPAddress `json:"ip_address"`
	Service   Service   `json:"service"`
	FirstTime time.Time `json:"first_time"`
	LastTime  time.Time `json:"last_time"`
}

func NewConnectedTo(ip IPAddress, service Service) ConnectedTo {
	return ConnectedTo{
		IPAddress: ip,
		Service:   service,
		FirstTime: Now(),
		LastTime:  Now(),
	}
}

func (c *ConnectedTo) ToCypher() (cypher string, params map[string]any) {
	cypher = `
		MATCH (ip:IPAddress {address: $ct_ip_address})
		MATCH (service:Service {name: $ct_service_name, port: $ct_service_port})
		MERGE (ip)-[ct:CONNECTED_TO]->(service)
		ON CREATE SET ct.times = 0, ct.fist_time = datetime($ct_first_time)
		SET ct.times = ct.times + 1, ct.last_time = datetime($ct_last_time)
		WITH ct
	`
	params = map[string]any{
		"ct_ip_address":   c.IPAddress.Address,
		"ct_service_name": c.Service.Name,
		"ct_service_port": c.Service.Port,
		"ct_first_time":   c.FirstTime.Format(time.RFC3339),
		"ct_last_time":    c.LastTime.Format(time.RFC3339),
	}
	return
}

type Username struct {
	Name     string    `json:"name"`
	Seen     int       `json:"seen"`
	FistSeen time.Time `json:"first_seen"`
	LastSeen time.Time `json:"last_seen"`
}

func NewUsername(name string) Username {
	return Username{
		Name:     name,
		FistSeen: Now(),
		LastSeen: Now(),
	}
}

func (u *Username) ToCypher() (cypher string, params map[string]any) {
	cypher = `
		MERGE (username:Username {name: $u_name})
		ON CREATE SET username.first_seen = datetime($u_first_seen), username.seen = 0
		SET username.last_seen = datetime($u_last_seen), username.seen = username.seen + 1
		WITH username
	`
	params = map[string]any{
		"u_name":       u.Name,
		"u_first_seen": u.FistSeen.Format(time.RFC3339),
		"u_last_seen":  u.LastSeen.Format(time.RFC3339),
	}
	return
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

func (a *AuthenticatedOn) ToCypher() (cypher string, params map[string]any) {
	cypher = `
		MATCH (username:Username {name: $ao_username_name})
		MATCH (service:Service {name: $ao_service_name, port: $ao_service_port})
		MERGE (username)-[a:AUTHENTICATED_ON]->(service)
		ON CREATE SET a.first_time = datetime($ao_first_time), a.failures = 0, a.successes = 0, a.times = 0
		SET a.last_time = datetime($ao_last_time), a.times = a.times + 1, 
		`
	if a.successful {
		cypher += `a.successes = a.successes + 1
		`
	} else {
		cypher += `a.failures = a.failures + 1
		`
	}
	cypher += `WITH a
	`

	params = map[string]any{
		"ao_username_name": a.Username.Name,
		"ao_service_name":  a.Service.Name,
		"ao_service_port":  a.Service.Port,
		"ao_first_time":    a.FirstTime.Format(time.RFC3339),
		"ao_last_time":     a.LastTime.Format(time.RFC3339),
	}
	return
}

type Cypher interface {
	ToCypher() (string, map[string]any)
}
