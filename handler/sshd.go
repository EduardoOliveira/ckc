package handler

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/EduardoOliveira/ckc/internal/opt"
	"github.com/EduardoOliveira/ckc/types"
	"github.com/elastic/go-grok"
)

type sshdParser struct {
	target types.Service
	groks  []*grok.Grok
	now    func() time.Time
}

func (h *sshdParser) Name() string {
	return "sshd_grok_parser"
}

func NewSSHDParser() sshdParser {
	rtn := sshdParser{
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

func (h *sshdParser) Parse(ctx context.Context, content string, parent types.ParsedEvent) (types.ParsedEvent, error) {
	parent.ServiceName = types.SSHDService
	matches, err := h.parseContent(content)
	if err != nil {
		return parent, fmt.Errorf("failed to parse content: %w", err)
	}
	parent.Service = types.Service{
		Name: "ssd",
		Host: parent.Hostname,
		Port: 22,
	}
	parent.IPAddress = types.IPAddress{
		Address: matches["system.auth.ip"],
	}
	parent.Username = types.Username{
		Name: matches["system.auth.user"],
	}

	parent.SSHDEvent = opt.Some(types.SSHDParsedEvent{
		Method:    matches["system.auth.ssh.method"],
		Signature: matches["system.auth.ssh.signature"],
		Result:    matches["system.auth.ssh.event"],
	})
	if strings.HasPrefix(matches["system.auth.ssh.event"], "Accepted") {
		parent.SSHDEvent.Value.Success = true
	}

	return parent, nil
}

func (h *sshdParser) parseContent(content string) (map[string]string, error) {
	bestMatch := map[string]string{}

	for _, g := range h.groks {
		matches, err := g.ParseString(content)
		if err != nil {
			fmt.Printf("Failed to parse log with grok pattern: %v\n", err)
			continue
		}
		if len(matches) > 0 && len(matches) > len(bestMatch) {
			bestMatch = matches
		}
	}
	if len(bestMatch) == 0 {
		return nil, fmt.Errorf("no grok pattern matched for content: %s", content)
	}
	return bestMatch, nil
}

/*
func (h *sshdParser) _Parse(content string) (types.IPAddress, []types.Cypher, error) {
	var ipAddress types.IPAddress
	var username types.Username

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
}*/
