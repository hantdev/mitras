package events

import "github.com/hantdev/mitras/pkg/events"

var _ events.Event = (*mqttEvent)(nil)

type mqttEvent struct {
	clientID  string
	operation string
	instance  string
}

func (me mqttEvent) Encode() (map[string]interface{}, error) {
	return map[string]interface{}{
		"client_id": me.clientID,
		"operation": me.operation,
		"instance":  me.instance,
	}, nil
}
