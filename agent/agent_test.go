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
	"encoding/json"
	"github.com/hashicorp/serf/serf"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestStartAgent(t *testing.T) {
	serfConfig := serf.DefaultConfig()
	serfConfig.NodeName = "test-agent"
	serfConfig.MemberlistConfig.BindPort = 7951

	a, err := StartAgent(serfConfig, nil, nil)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer a.Shutdown()
	time.Sleep(10 * time.Millisecond)

	for _, m := range a.serf.Members() {
		if m.Name != "test-agent" {
			t.Fatalf("Wrong member name. Expected 'test-agent', but got '%s'", m.Name)
		}
	}
}

func TestEventHandlers(t *testing.T) {
	serfConfig := serf.DefaultConfig()
	serfConfig.NodeName = "test-agent"
	serfConfig.MemberlistConfig.BindPort = 7952

	var me *serf.MemberEvent
	var ue *serf.UserEvent

	memberEventHandler := func(e serf.Event) {
		if e.EventType() == serf.EventMemberJoin {
			tmp := e.(serf.MemberEvent)
			me = &tmp
		}
	}
	userEventHandler := func(e serf.Event) {
		if e.EventType() == serf.EventUser {
			tmp := e.(serf.UserEvent)
			ue = &tmp
		}
	}

	a, err := StartAgent(serfConfig, nil, []HandleEvent{memberEventHandler, userEventHandler})
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer a.Shutdown()
	time.Sleep(20 * time.Millisecond)

	err = a.serf.UserEvent("custom-event", []byte("my payload"), false)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	time.Sleep(10 * time.Millisecond)

	if me == nil {
		t.Fatal("Event handler hasn't receive a member join event")
	}
	if ue == nil {
		t.Fatal("Event handler hasn't receive a user event")
	}

	if me.Type != serf.EventMemberJoin {
		t.Fatalf("Wrong event type. Expected 'member-join', but got '%s'", me.Type)
	}

	if ue.Name != "custom-event" {
		t.Fatalf("Wrong event name. Expected 'custom-event', but got '%s'", ue.Name)
	}
	if string(ue.Payload) != "my payload" {
		t.Fatalf("Wrong event payload. Expected 'my payload', but got '%s'", ue.Payload)
	}
}

func TestLoadKeyringFile(t *testing.T) {
	td, err := ioutil.TempDir("", "serf")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.RemoveAll(td)

	keyringFile := filepath.Join(td, "keyring.json")

	encodedKeys, err := json.Marshal([]string{
		"enjTwAFRe4IE71bOFhirzQ==",
		"csT9mxI7aTf9ap3HLBbdmA==",
		"noha2tVc0OyD/2LtCBoAOQ==",
	})
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if err := ioutil.WriteFile(keyringFile, encodedKeys, 0600); err != nil {
		t.Fatalf("err: %s", err)
	}

	a := &Agent{config: serf.DefaultConfig()}
	err = a.loadKeyringFile(keyringFile)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	totalLoadedKeys := len(a.config.MemberlistConfig.Keyring.GetKeys())
	if totalLoadedKeys != 3 {
		t.Fatalf("Expected to load 3 keys but got %d", totalLoadedKeys)
	}
}

func TestAgent_UserEvent(t *testing.T) {
	serfConfig := serf.DefaultConfig()
	serfConfig.NodeName = "test-agent"
	serfConfig.MemberlistConfig.BindPort = 7953

	var ue *serf.UserEvent

	userEventHandler := func(e serf.Event) {
		if e.EventType() == serf.EventUser {
			tmp := e.(serf.UserEvent)
			ue = &tmp
		}
	}

	a, err := StartAgent(serfConfig, nil, []HandleEvent{userEventHandler})
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer a.Shutdown()
	time.Sleep(20 * time.Millisecond)

	err = a.UserEvent("custom-event", []byte("my payload"), false)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	time.Sleep(10 * time.Millisecond)

	if ue == nil {
		t.Fatal("Event handler hasn't receive a user event")
	}

	if ue.Name != "custom-event" {
		t.Fatalf("Wrong event name. Expected 'custom-event', but got '%s'", ue.Name)
	}
	if string(ue.Payload) != "my payload" {
		t.Fatalf("Wrong event payload. Expected 'my payload', but got '%s'", ue.Payload)
	}
	if ue.Coalesce {
		t.Fatalf("Wrong coalesce. Expected 'false', but got '%t'", ue.Coalesce)
	}
}
