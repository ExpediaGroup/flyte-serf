# flyte-serf 

![Build Status](https://travis-ci.org/HotelsDotCom/flyte-serf.svg?branch=master)
[![Docker Stars](https://img.shields.io/docker/stars/hotelsdotcom/flyte-serf.svg)](https://hub.docker.com/r/hotelsdotcom/flyte-serf)
[![Docker Pulls](https://img.shields.io/docker/pulls/hotelsdotcom/flyte-serf.svg)](https://hub.docker.com/r/hotelsdotcom/flyte-serf)

A Serf pack for Flyte which will join an existing Serf cluster and allow you to raise and respond to events from your Flyte flows.

It provides two triggers `UserEvent`, `MemberEvent` and one command `SendEvent`

## ENV Configuration

| env. variable     | default                      |description                      |
|-------------------|------------------------------|---------------------------------|
| SERF_NODE_NAME    | flyte-serf                   | The name of this node in the cluster. |
| SERF_ADVERTISE    |                              | Address that we advertise to other nodes in the cluster. By default, the bind address is advertised. |
| SERF_JOIN_HOSTS   |                              | Address of another agent to join upon starting up. This can be comma delimited to specify multiple agents to join. |
| SERF_KEYRING_FILE |                              | Specifies a file to load keyring data from. |
| FLYTE_API         |                              | The FLYTE API endpoint to use. |

## Events

### UserEvent
This event is invoked when Serf agent receives user event. Event output is a map of strings 
```
{
    "name": "...user event name...", 
    "payload": "...user event payload..."
}
```

### MemberEvent
This event is invoked when Serf agent receives member event: join, leave or failed. This event is disabled currently.
Event output is a map of strings
```
{
    "name": "...member event name...",
    "memberName": "...Serf node name...",
    "memberAddr": "...Serf node address..."
}
```

### EventSent
This event is emitted on the successful invocation of the `SendEvent` command. It's payload is:
```
{
    "name": "...user event name...", 
    "payload": "...user event payload...", 
    "coalesce": "...coalesce flag..."
}
```

### SendEventFailed
This event is emitted on the unsuccessful invocation of the `SendEvent` command. It's payload is an error message.

## Commands

### SendEvent
This command sends user event to a serf cluster. Command's input is 
```
{
    "name": "...user event name...", 
    "payload": "...user event payload...", 
    "coalesce": "...coalesce flag..."
}
```

This command can output two event types: `EventSent` and `SendEventFailed`




## Build and run

### GO

To build and run from the command line:
* Clone this repo
* Run `dep ensure` (must have [dep](https://github.com/golang/dep) installed )
* Run `go build`
* FLYTE_API="FLYTE_API_URL" ./flyte-serf
    
### Docker

    docker build -t flyte-serf:latest .
    docker run --rm flyte-serf:latest
