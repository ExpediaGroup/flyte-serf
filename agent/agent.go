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

package agent

import (
	"fmt"
	"github.com/hashicorp/serf/cmd/serf/command/agent"
	"github.com/hashicorp/serf/serf"
	"log"
	"net"
	"sync"
	"time"
)

const (
	// gracefulTimeout controls how long we wait before forcefully terminating
	gracefulTimeout = 3 * time.Second
)

// Agent starts and manages a Serf instance, adding some niceties
// on top of Serf such as invoking EventHandlers when events occur.
type Agent struct {
	// Stores serf configuration
	config *serf.Config

	// This is the underlying Serf we are wrapping
	serf *serf.Serf

	// eventCh is used for Serf to deliver events on
	eventCh chan serf.Event

	// eventHandlers is the registered handlers for events
	eventHandlers []HandleEvent

	// shutdownCh is used for shutdowns
	shutdown     bool
	shutdownCh   chan struct{}
	shutdownLock sync.Mutex
}

type HandleEvent func(e serf.Event)

func StartAgent(conf *serf.Config, joinHosts []string, eventHandlers []HandleEvent) (*Agent, error) {
	a := &Agent{
		config:        conf,
		eventCh:       make(chan serf.Event, 64),
		shutdownCh:    make(chan struct{}),
		eventHandlers: eventHandlers,
	}
	a.config.EventCh = a.eventCh

	// Load in a keyring file if provided
	if conf.KeyringFile != "" {
		err := a.loadKeyringFile(conf.KeyringFile)
		if err != nil {
			return nil, fmt.Errorf("Failed to load keyring file: %s", err)
		}
	}

	s, err := serf.Create(a.config)
	if err != nil {
		return nil, fmt.Errorf("Failed to start Serf: %s", err)
	}
	a.serf = s

	log.Println("[INFO] Started Serf agent!")
	log.Printf("[INFO] Serf Node name: '%s'", conf.NodeName)
	bindAddr := (&net.TCPAddr{IP: net.ParseIP(conf.MemberlistConfig.BindAddr), Port: conf.MemberlistConfig.BindPort}).String()
	log.Printf("[INFO] Serf Bind addr: '%s'", bindAddr)
	if conf.MemberlistConfig.AdvertiseAddr != "" {
		advertiseAddr := (&net.TCPAddr{IP: net.ParseIP(conf.MemberlistConfig.AdvertiseAddr), Port: conf.MemberlistConfig.AdvertisePort}).String()
		log.Printf("[INFO] Serf Advertise addr: '%s'", advertiseAddr)
	}

	if joinHosts != nil && len(joinHosts) > 0 {
		a.Join(joinHosts, true)
	}

	// Start ingesting events for Serf
	if a.eventHandlers != nil {
		go a.handleSerfEvents()
	}

	return a, nil
}

// Join asks the Serf instance to join. See the Serf.Join function.
func (a *Agent) Join(addrs []string, ignoreOld bool) {
	log.Printf("[INFO] Joining Serf cluster through hosts: %v ignore old: %v", addrs, ignoreOld)
	n, err := a.serf.Join(addrs, ignoreOld)
	if n > 0 {
		log.Printf("[INFO] Serf joined %d nodes", n)
	}
	if err != nil {
		//Exit in case cluster join fails
		log.Fatalf("[ERROR] Serf couldn't join cluster: %s", err)
	}
}

func (a *Agent) UserEvent(name string, payload []byte, coalesce bool) error {
	return a.serf.UserEvent(name, payload, coalesce)
}

// handleSerfEvents is used to handle events from the Serf cluster
func (a *Agent) handleSerfEvents() {
	for {
		select {
		case e := <-a.eventCh:
			for _, eh := range a.eventHandlers {
				go eh(e)
			}

		case <-a.shutdownCh:
			return
		}
	}
}

// Leave executes graceful shutdown of the agent and its processes
func (a *Agent) Leave() {
	if a.serf != nil {
		log.Println("[INFO] Requesting graceful leave from Serf")

		gracefulCh := make(chan error)
		go func() {
			gracefulCh <- a.serf.Leave()
		}()

		select {
		case <-time.After(gracefulTimeout):
			log.Println("[WARN] Timeout while waiting for graceful leave")
		case err := <-gracefulCh:
			if err != nil {
				log.Printf("[ERROR] Error while leaving Serf: %s", err)
			} else {
				log.Println("[INFO] Gracefully left Serf")
			}
		case <-a.shutdownCh:
			log.Println("[WARN] Serf agent is already shutdown")
		}
	}
}

// Shutdown closes this agent and all of its processes. Should be preceded by a Leave for a graceful shutdown.
func (a *Agent) Shutdown() error {
	a.shutdownLock.Lock()
	defer a.shutdownLock.Unlock()

	if a.shutdown {
		return nil
	}

	if a.serf != nil {
		log.Println("[INFO] Requesting Serf shutdown")
		if err := a.serf.Shutdown(); err != nil {
			return err
		}
	}

	log.Println("[INFO] Serf agent shutdown complete")
	a.shutdown = true
	close(a.shutdownCh)
	return nil
}

// ShutdownCh returns a channel that can be selected to wait
// for the agent to perform a shutdown.
func (a *Agent) ShutdownCh() <-chan struct{} {
	return a.shutdownCh
}

// loadKeyringFile will load a keyring out of a file
// We are using Serf command agent to load file and then get Keyring from config
// It is not possible to have our own implementation as serf config is using vendor packages for memberlist.Keyring
// a.config.MemberlistConfig.Keyring is of this type
// github.com/hashicorp/serf/vendor/github.com/hashicorp/memberlist.Keyring
// and not simply of this type: github.com/hashicorp/memberlist.Keyring
// Another option would be just to setup a a.config.MemberlistConfig.SecretKey which is just a byte slice
func (a *Agent) loadKeyringFile(keyringFile string) error {
	serfConf := serf.DefaultConfig()
	agentConf := agent.DefaultConfig()
	agentConf.KeyringFile = keyringFile
	_, err := agent.Create(agentConf, serfConf, nil)
	if err != nil {
		return fmt.Errorf("Error loading keyring file: %s", err)
	}
	a.config.MemberlistConfig.Keyring = serfConf.MemberlistConfig.Keyring
	// Success!
	return nil
}
