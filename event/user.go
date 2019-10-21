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
)

type UserEventHandler struct {
	pack flyte.Pack
}

func (h *UserEventHandler) SetPack(pack flyte.Pack) {
	h.pack = pack
}

func (h *UserEventHandler) HandleUserEvent(e serf.Event) {
	switch e.EventType() {
	case serf.EventUser:
		ue := e.(serf.UserEvent)
		args := map[string]string{}
		args["name"] = ue.Name
		args["payload"] = string(ue.Payload)
		log.Printf("[INFO] Received event: %s %s", args["name"], args["payload"])
		if err := h.pack.SendEvent(flyte.Event{UserEventDef, args}); err != nil {
			log.Print(err)
		}
	}
}
