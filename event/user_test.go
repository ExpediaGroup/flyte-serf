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

package event

import (
	"github.com/hashicorp/serf/serf"
	"testing"
)

func TestHandleUserEvent(t *testing.T) {
	pack := &mockPack{}
	pack.Start()

	h := UserEventHandler{pack}
	h.HandleUserEvent(serf.UserEvent{
		Name:    "custom-event",
		Payload: []byte("my payload"),
	})

	select {
	case out := <-pack.events:
		if out.EventDef != UserEventDef {
			t.Fatalf("Wrong event.\nWant: %s\nGot : %s", UserEventDef.Name, out.EventDef.Name)
		}
		args := out.Payload.(map[string]string)
		if args["name"] != "custom-event" {
			t.Errorf("Wrong name. Expected: 'custom-event', but got '%s'", args["name"])
		}
		if args["payload"] != "my payload" {
			t.Errorf("Wrong payload. Expected: 'my payload', but got '%s'", args["payload"])
		}
	default:
		t.Fatal("Should handle user event")
	}
}

func TestUserEventHandler_ShouldNotHandleMemberEvents(t *testing.T) {
	pack := &mockPack{}
	pack.Start()

	h := UserEventHandler{pack}
	h.HandleUserEvent(serf.MemberEvent{})

	select {
	case <-pack.events:
		t.Error("Should not handle member event")
	default:
	}
}
