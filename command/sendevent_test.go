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

package command

import (
	"encoding/json"
	"github.com/hashicorp/serf/serf"
	"github.com/HotelsDotCom/flyte-serf/agent"
	"github.com/HotelsDotCom/flyte-serf/event"
	"testing"
	"time"
)

func TestSendEventHandler(t *testing.T) {
	serfConfig := serf.DefaultConfig()
	serfConfig.MemberlistConfig.BindPort = 7954

	var ue *serf.UserEvent

	userEventHandler := func(e serf.Event) {
		if e.EventType() == serf.EventUser {
			tmp := e.(serf.UserEvent)
			ue = &tmp
		}
	}

	a, err := agent.StartAgent(serfConfig, nil, []agent.HandleEvent{userEventHandler})
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer a.Shutdown()
	time.Sleep(20 * time.Millisecond)

	command := SendEventCommand(a)

	out := command.Handler(json.RawMessage(`{"name":"custom-event", "payload": "my payload", "coalesce":false}`))

	time.Sleep(10 * time.Millisecond)

	if out.EventDef != event.EventSentDef {
		t.Fatalf("Wrong event.\nWant: %s\nGot : %s", event.EventSentDef.Name, out.EventDef.Name)
	}

	if ue == nil {
		t.Fatal("Event handler hasn't received a user event")
	}

	if ue.Name != "custom-event" {
		t.Fatalf("Wrong event name. Expected 'custom-event', but got '%s'", ue.Name)
	}
	if string(ue.Payload) != "my payload" {
		t.Fatalf("Wrong event payload. Expected 'my payload', but got '%s'", ue.Payload)
	}
	if ue.Coalesce {
		t.Fatalf("Wrong coalesce. Expected 'true', but got '%t'", ue.Coalesce)
	}
	a.Leave()
}
