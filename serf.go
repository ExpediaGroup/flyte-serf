/*
Copyright (C) 2018 Expedia Group.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"github.com/HotelsDotCom/flyte-serf/agent"
	"github.com/hashicorp/serf/serf"
	"log"
	"os"
	"strconv"
	"strings"
)

func initializeSerfAgent(eventHandlers []agent.HandleEvent) (*agent.Agent, error) {
	conf := serf.DefaultConfig()
	conf.Init()

	conf.NodeName = "flyte-serf"
	if os.Getenv("SERF_NODE_NAME") != "" {
		conf.NodeName = os.Getenv("SERF_NODE_NAME")
	}

	conf.Tags["role"] = "flyte-serf"

	if os.Getenv("SERF_KEYRING_FILE") != "" {
		conf.KeyringFile = os.Getenv("SERF_KEYRING_FILE")
	}

	if os.Getenv("SERF_ADVERTISE") != "" {
		advertise := strings.Split(os.Getenv("SERF_ADVERTISE"), ":")
		conf.MemberlistConfig.AdvertiseAddr = advertise[0]
		if len(advertise) > 1 {
			p, err := strconv.Atoi(advertise[1])
			if err != nil {
				log.Fatalf("[ERROR] Unable to resolve Serf advertise port number: %s", err)
			} else {
				conf.MemberlistConfig.AdvertisePort = p
			}
		}
	}

	joinHosts := []string{}
	if os.Getenv("SERF_JOIN_HOSTS") != "" {
		joinHosts = strings.Split(os.Getenv("SERF_JOIN_HOSTS"), ",")
	} else {
		log.Println("[WARN] Serf agent won't join any cluster at the startup: SERF_JOIN_HOSTS is not specified")
	}

	return agent.StartAgent(conf, joinHosts, eventHandlers)
}
