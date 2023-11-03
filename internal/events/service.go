// Licensed to The Moov Authors under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. The Moov Authors licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package events

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/moov-io/ach"
	"github.com/moov-io/achgateway/internal/service"
	"github.com/moov-io/achgateway/pkg/compliance"
	"github.com/moov-io/achgateway/pkg/models"
	"github.com/moov-io/base/log"
)

type Emitter interface {
	Send(ctx context.Context, evt models.Event) error
}

func NewEmitter(logger log.Logger, cfg *service.EventsConfig) (Emitter, error) {
	if cfg == nil {
		return &MockEmitter{}, nil
	}
	if cfg.Stream != nil {
		return newStreamService(logger, cfg.Transform, cfg.Stream)
	}
	if cfg.Webhook != nil {
		return newWebhookService(logger, cfg.Transform, cfg.Webhook)
	}
	return nil, errors.New("unknown events config")
}

type MockEmitter struct {
	mu   sync.Mutex
	sent []models.Event
}

func (e *MockEmitter) Send(_ context.Context, evt models.Event) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Compute the full cycle of events like real implementations do
	bs, err := compliance.Protect(nil, evt)
	if err != nil {
		return fmt.Errorf("mock emitter - protect: %w", err)
	}
	bs, err = compliance.Reveal(nil, bs)
	if err != nil {
		return fmt.Errorf("mock emitter - reveal: %w", err)
	}

	var validateOpts *ach.ValidateOpts
	switch event := evt.Event.(type) {
	case models.ReconciliationEntry:
		validateOpts = &ach.ValidateOpts{
			AllowMissingFileControl:    true,
			AllowMissingFileHeader:     true,
			AllowUnorderedBatchNumbers: true,
		}
	case models.ReconciliationFile:
		validateOpts = event.File.GetValidation()
	}

	found, err := models.ReadWithOpts(bs, validateOpts)
	if err != nil {
		return fmt.Errorf("mock emitter - read: %w", err)
	}

	if found != nil {
		e.sent = append(e.sent, *found)
	}

	return nil
}

func (e *MockEmitter) Sent() []models.Event {
	e.mu.Lock()
	defer e.mu.Unlock()

	out := make([]models.Event, len(e.sent))
	copy(out, e.sent)
	return out
}
