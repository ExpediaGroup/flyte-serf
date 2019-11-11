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
	"github.com/HotelsDotCom/flyte-client/flyte"
	"github.com/hashicorp/serf/serf"
	"log"
	"net"
)

type MemberEventHandler struct {
	pack flyte.Pack
}

func (h *MemberEventHandler) SetPack(pack flyte.Pack) {
	h.pack = pack
}

func (h MemberEventHandler) HandleMemberEvent(e serf.Event) {
	switch e.EventType() {
	case serf.EventMemberJoin, serf.EventMemberLeave, serf.EventMemberFailed:
		me := e.(serf.MemberEvent)
		args := map[string]string{}
		args["name"] = me.Type.String()
		for _, m := range me.Members {
			args["memberName"] = m.Name
			args["memberAddr"] = (&net.TCPAddr{IP: m.Addr, Port: int(m.Port)}).String()
			log.Printf("[INFO] Received member event: %s: %s %s", args["name"], args["memberName"], args["memberAddr"])

			if err := h.pack.SendEvent(flyte.Event{MemberEventDef, args}); err != nil {
				log.Print(err)
			}
		}
	}
}
