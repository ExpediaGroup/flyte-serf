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
	"errors"
	"log"
	"github.com/HotelsDotCom/flyte-client/flyte"
	"github.com/HotelsDotCom/flyte-serf/agent"
	"github.com/HotelsDotCom/flyte-serf/event"
)

func SendEventCommand(serfAgent *agent.Agent) flyte.Command {
	c := sendEventCommand{
		serfAgent: serfAgent,
	}
	return flyte.Command{
		Name:         "SendEvent",
		OutputEvents: []flyte.EventDef{event.EventSentDef},
		Handler:      c.sendEventHandler,
	}
}

type sendEventCommand struct {
	serfAgent *agent.Agent
}

type userEvent struct {
	Name     string `json:"name"`
	Payload  string `json:"payload"`
	Coalesce bool   `json:"coalesce"`
}

func (e userEvent) getPayload() []byte {
	if e.Payload == "" {
		return nil
	}
	return []byte(e.Payload)
}

func (c sendEventCommand) sendEventHandler(input json.RawMessage) flyte.Event {
	var e userEvent
	if err := json.Unmarshal(input, &e); err != nil {
		log.Printf("[ERROR] Serf couldn't unmarshal input: %s\n", err)
		return event.NewSendEventFailedEvent(err)
	}

	if e.Name == "" {
		err := errors.New("Can't send user event: Event name is not provided")
		return event.NewSendEventFailedEvent(err)
	}

	log.Printf("[INFO] Sending event: '%s' with payload: '%s' and coalesce: '%v'", e.Name, e.Payload, e.Coalesce)
	err := c.serfAgent.UserEvent(e.Name, e.getPayload(), e.Coalesce)
	if err != nil {
		log.Printf("[ERROR] Serf couldn't send user event: %s\n", err)
		return event.NewSendEventFailedEvent(err)
	}

	return event.NewEventSentEvent(e)
}
