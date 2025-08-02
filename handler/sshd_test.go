package handler

import (
	"context"
	"testing"

	"github.com/EduardoOliveira/ckc/internal/time_help"
	"github.com/EduardoOliveira/ckc/types"
	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/stretchr/testify/assert"
)

func TestParseSSHDLog(t *testing.T) {
	// Create a new SSHD handler
	handler := NewSSHDParser()
	handler.now = time_help.Now
	types.Now = time_help.Now
	/*
	   	Feb 20 21:54:44 localhost sshd[3402]: Accepted publickey for vagrant from 10.0.2.2 port 63673 ssh2: RSA 39:33:99:e9:a0:dc:f2:33:a3:e5:72:3b:7c:3a:56:84
	   Feb 21 00:13:35 localhost sshd[7483]: Accepted password for vagrant from 192.168.33.1 port 58803 ssh2
	   Feb 21 08:35:22 localhost sshd[5774]: Failed password for root from 116.31.116.24 port 29160 ssh2
	   Feb 21 19:19:26 localhost sshd[16153]: Failed password for invalid user aurelien from 142.0.45.14 port 52772 ssh2
	   Feb 21 21:56:12 localhost sshd[3430]: Invalid user test from 10.0.2.2
	*/
	testCases := []struct {
		name     string
		log      string
		expected []types.ParsedEvent
	}{
		{
			name: "Valid SSHD log with user",
			log:  `Accepted publickey for vagrant from 10.0.2.2 port 63673 ssh2: RSA 39:33:99:e9:a0:dc:f2:33:a3:e5:72:3b:7c:3a:56:84`,
		},
		{
			name: "Valid SSHD log with password",
			log:  `Accepted password for vagrant from 192.168.33.1 port 58803 ssh2`,
		},
		{
			name: "Failed SSHD log with root user",
			log:  `Failed password for root from 116.31.116.24 port 29160 ssh2`,
		},
		{
			name: "Failed SSHD log with invalid user",
			log:  `Failed password for invalid user aurelien from 142.0.45.14 port 52772 ssh2`,
		},
		{
			name: "Invalid user SSHD log",
			log:  `Invalid user test from 10.0.2.2`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := handler.Parse(context.Background(), tc.log, types.ParsedEvent{
				Ingestion: time_help.Now(),
			})
			assert.NoError(t, err)
			snaps.MatchSnapshot(t, result)
		})
	}
}
