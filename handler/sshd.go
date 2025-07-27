package handler

import (
	"fmt"
	"strings"
	"time"

	"github.com/EduardoOliveira/ckc/internal/ptr"
	"github.com/EduardoOliveira/ckc/types"
	"github.com/elastic/go-grok"
)

type sshdHandler struct {
	target types.Service
	groks  []*grok.Grok
	now    func() time.Time
}

func NewSSHDHandler() sshdHandler {
	rtn := sshdHandler{
		groks: make([]*grok.Grok, 2),
		now:   time.Now,
		target: types.Service{
			Name: "ssh",
			Port: 22,
		},
	}
	rtn.groks[0] = grok.New()
	err := rtn.groks[0].Compile(`%{DATA:system.auth.ssh.event} %{DATA:system.auth.ssh.method} for (invalid user )?%{DATA:system.auth.user} from %{IPORHOST:system.auth.ip} port %{NUMBER:system.auth.port} ssh2(: %{GREEDYDATA:system.auth.ssh.signature})?`, true)
	if err != nil {
		panic(fmt.Sprintf("Failed to compile grok pattern: %v", err))
	}

	rtn.groks[1] = grok.New()
	err = rtn.groks[1].Compile(`%{DATA:system.auth.ssh.event} user %{DATA:system.auth.user} from %{IPORHOST:system.auth.ip}`, true)
	if err != nil {
		panic(fmt.Sprintf("Failed to compile grok pattern: %v", err))
	}

	return rtn
}

func (h *sshdHandler) Parse(content string) (types.IPAddress, []types.Cypher, error) {
	var ipAddress types.IPAddress
	var username types.Username
	var success bool
	bestMatch := 0
	for _, g := range h.groks {
		matches, err := g.ParseString(content)
		if err != nil {
			fmt.Printf("Failed to parse log with grok pattern: %v\n", err)
			continue
		}
		if len(matches) > 0 && len(matches) > bestMatch {
			bestMatch = len(matches)
			fmt.Println("Matched Grok pattern:", matches)
			ipAddress = types.NewIPAddress(matches["system.auth.ip"])
			username = types.NewUsername(matches["system.auth.user"])
			if strings.HasPrefix(matches["system.auth.ssh.event"], "Accepted") {
				success = true
			} else {
				success = false
			}
		}
	}
	if bestMatch > 0 {
		return ipAddress, []types.Cypher{
			&ipAddress,
			&username,
			ptr.To(types.NewWithUsername(ipAddress, username)),
			&h.target,
			ptr.To(types.NewAuthenticatedOn(username, h.target, success)),
			ptr.To(types.NewConnectedTo(ipAddress, h.target)),
		}, nil
	}
	return ipAddress, []types.Cypher{}, nil
}
