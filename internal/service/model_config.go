// generated-from:9e0782b937278abaee17ffb9be40bb7928f6d9aeac4d35aa713f071163fd474c DO NOT REMOVE, DO UPDATE

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

package service

import (
	"fmt"

	"github.com/moov-io/base/database"
	"github.com/moov-io/base/log"
	"github.com/moov-io/base/telemetry"
)

type GlobalConfig struct {
	ACHGateway Config
}

type Config struct {
	Logger    log.Logger `json:"-"`
	Clients   *ClientConfig
	Database  database.DatabaseConfig
	Telemetry telemetry.Config
	Admin     Admin
	Inbound   Inbound
	Events    *EventsConfig
	Sharding  Sharding
	Upload    UploadAgents
	Errors    ErrorAlerting
}

func (cfg *Config) Validate() error {
	if err := cfg.Admin.Validate(); err != nil {
		return fmt.Errorf("admin: %v", err)
	}
	if err := cfg.Inbound.Validate(); err != nil {
		return fmt.Errorf("inbound: %v", err)
	}
	if err := cfg.Events.Validate(); err != nil {
		return fmt.Errorf("events: %v", err)
	}
	if err := cfg.Sharding.Validate(); err != nil {
		return fmt.Errorf("sharding: %v", err)
	}
	if err := cfg.Upload.Validate(); err != nil {
		return fmt.Errorf("upload: %v", err)
	}
	if err := cfg.Errors.Validate(); err != nil {
		return fmt.Errorf("errors: %v", err)
	}
	return nil
}
