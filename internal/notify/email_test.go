// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package notify

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/moov-io/achgateway/internal/service"

	"github.com/stretchr/testify/require"

	"github.com/moov-io/ach"
)

func TestEmailSend(t *testing.T) {
	dep := spawnMailslurp(t)

	cfg := &service.Email{
		ID:   "testing",
		From: "noreply@moov.io",
		To: []string{
			"jane@company.com",
		},
		ConnectionURI: fmt.Sprintf("smtps://test:test@localhost:%s/?insecure_skip_verify=true", dep.SMTPPort()),
		CompanyName:   "Moov",
	}

	dialer, err := setupGoMailClient(cfg)
	require.NoError(t, err)
	// Enable SSL for our test container, this breaks if set for production SMTP server.
	// GMail fails to connect if we set this.
	dialer.SSL = strings.HasPrefix(cfg.ConnectionURI, "smtps://")

	msg := &Message{
		Direction: Upload,
		Filename:  "20200529-131400.ach",
		File:      ach.NewFile(),
	}

	body, err := marshalEmail(cfg, msg)
	require.NoError(t, err)

	ctx := context.Background()
	if err := sendEmail(ctx, cfg, dialer, msg.Filename, body); err != nil {
		t.Fatal(err)
	}

	dep.Close() // remove container after successful tests
}

func TestEmail__marshalDefaultTemplate(t *testing.T) {
	f, err := ach.ReadFile(filepath.Join("..", "..", "testdata", "ppd-debit.ach"))
	require.NoError(t, err)

	tests := []struct {
		desc      string
		msg       *Message
		firstLine string
	}{
		{"upload with hostname", &Message{Direction: Upload, File: f, Filename: "20200529-131400.ach", Hostname: "ftp.bank.com:3294"},
			"A file has been uploaded to ftp.bank.com:3294 - 20200529-131400.ach"},
		{"upload with no hostname", &Message{Direction: Upload, File: f, Filename: "20200529-131400.ach"},
			"A file has been uploaded - 20200529-131400.ach"},
		{"download with hostname", &Message{Direction: Download, File: f, Filename: "20200529-131400.ach", Hostname: "138.34.204.3"},
			"A file has been downloaded from 138.34.204.3 - 20200529-131400.ach"},
		{"download", &Message{Direction: Download, File: f, Filename: "20200529-131400.ach"},
			"A file has been downloaded - 20200529-131400.ach"},
	}

	cfg := &service.Email{
		CompanyName: "Moov",
	}

	for _, test := range tests {
		contents, err := marshalEmail(cfg, test.msg)
		if err != nil {
			t.Fatal(err)
		}

		if testing.Verbose() {
			t.Log(contents)
		}

		require.Contains(t, contents, test.firstLine, "Test: "+test.desc)
		require.Contains(t, contents, "Moov")
		require.Contains(t, contents, `Debits:  $105.00`, "Test: "+test.desc)
		require.Contains(t, contents, `Credits: $0.00`, "Test: "+test.desc)
		require.Contains(t, contents, `Batches: 1`, "Test: "+test.desc)
		require.Contains(t, contents, `Total Entries: 1`, "Test: "+test.desc)
	}
}
