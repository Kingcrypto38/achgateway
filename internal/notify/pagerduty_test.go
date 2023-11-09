// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package notify

import (
	"context"
	"os"
	"testing"

	"github.com/moov-io/ach"
	"github.com/moov-io/achgateway/internal/service"

	"github.com/stretchr/testify/require"
)

func testPagerDutyClient(t *testing.T) *PagerDuty {
	t.Helper()

	cfg := &service.PagerDuty{
		ID:         "testing",
		ApiKey:     os.Getenv("PAGERDUTY_API_KEY"),
		From:       "adam@moov.io",
		ServiceKey: "PM8YUZY", // testing
	}
	if cfg.ApiKey == "" {
		t.Skip("missing PagerDuty api key")
	}

	client, err := NewPagerDuty(cfg)
	require.NoError(t, err)

	return client
}

func TestPagerDuty(t *testing.T) {
	pd := testPagerDutyClient(t)

	if err := pd.Ping(); err != nil {
		t.Fatal(err)
	}

	file := ach.NewFile()
	ctx := context.Background()

	if err := pd.Info(ctx, &Message{
		Direction: Download,
		Filename:  "20200529-140002-1.ach",
		File:      file,
	}); err != nil {
		t.Fatal(err)
	}

	if err := pd.Critical(ctx, &Message{
		Direction: Upload,
		Filename:  "20200529-140002-2.ach",
		File:      file,
	}); err != nil {
		t.Fatal(err)
	}
}
