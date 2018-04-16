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
	"net"
	"testing"
)

func TestHandleMemberEvent(t *testing.T) {
	pack := &mockPack{}
	pack.Start()

	h := MemberEventHandler{pack}
	h.HandleMemberEvent(serf.MemberEvent{
		Type: serf.EventMemberJoin,
		Members: []serf.Member{
			{
				Name: "flyte-serf",
				Addr: net.IPv4(127, 0, 0, 1),
				Port: 7955,
			}},
	})

	select {
	case out := <-pack.events:
		if out.EventDef != MemberEventDef {
			t.Fatalf("Wrong event.\nWant: %s\nGot : %s", MemberEventDef.Name, out.EventDef.Name)
		}
		args := out.Payload.(map[string]string)
		if args["name"] != "member-join" {
			t.Errorf("Wrong name. Expected: 'member-join', but got '%s'", args["name"])
		}
		if args["memberName"] != "flyte-serf" {
			t.Errorf("Wrong memberName. Expected: 'flyte-serf', but got '%s'", args["memberName"])
		}
		if args["memberAddr"] != "127.0.0.1:7955" {
			t.Errorf("Wrong memberAddr. Expected: '127.0.0.1:7955', but got '%s'", args["memberAddr"])
		}
	default:
		t.Fatal("Should handle member event")
	}
}

func TestHandleMemberEvent_ShouldNotHandleUserEvent(t *testing.T) {
	pack := &mockPack{}
	pack.Start()

	h := MemberEventHandler{pack}
	h.HandleMemberEvent(serf.UserEvent{})

	select {
	case <-pack.events:
		t.Error("Should not handled user event")
	default:
	}
}
