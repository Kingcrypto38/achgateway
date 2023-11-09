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

package pipeline

import (
	"context"
	"fmt"

	"github.com/moov-io/achgateway/internal/events"
	"github.com/moov-io/achgateway/internal/files"
	"github.com/moov-io/achgateway/internal/incoming/stream"
	"github.com/moov-io/achgateway/internal/service"
	"github.com/moov-io/achgateway/internal/shards"
	"github.com/moov-io/achgateway/pkg/models"
	"github.com/moov-io/base/log"
)

func Start(
	ctx context.Context,
	logger log.Logger,
	cfg *service.Config,
	shardRepository shards.Repository,
	fileRepository files.Repository,
	httpFiles stream.Subscription,
) (*FileReceiver, error) {

	eventEmitter, err := events.NewEmitter(logger, cfg.Events)
	if err != nil {
		return nil, fmt.Errorf("pipeline: error creating event emitter: %v", err)
	}

	// register each shard's aggregator
	shardAggregators := make(map[string]*aggregator)
	for i := range cfg.Sharding.Shards {
		xfagg, err := newAggregator(logger, eventEmitter, cfg.Sharding.Shards[i], cfg.Upload, cfg.Errors)
		if err != nil {
			return nil, fmt.Errorf("problem starting shard=%s: %v", cfg.Sharding.Shards[i].Name, err)
		}

		go xfagg.Start(ctx)

		shardName := cfg.Sharding.Shards[i].Name
		shardAggregators[shardName] = xfagg
		initializeShardMetrics(shardName)
	}

	// register our fileReceiver and start it
	var transformConfig *models.TransformConfig
	if cfg.Inbound.Kafka != nil && cfg.Inbound.Kafka.Transform != nil {
		transformConfig = cfg.Inbound.Kafka.Transform
	}
	receiver, err := newFileReceiver(logger, cfg, eventEmitter, shardRepository, shardAggregators, fileRepository, httpFiles, transformConfig)
	if err != nil {
		return nil, err
	}
	go receiver.Start(ctx)

	return receiver, nil
}
