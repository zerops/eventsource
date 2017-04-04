package eventsource

import (
	"encoding/json"
	"reflect"
)

type Serializer interface {
	Serialize(event Event) (Record, error)
	Deserialize(record Record) (Event, error)
}

type jsonEvent struct {
	Type string          `json:"t"`
	Data json.RawMessage `json:"d"`
}

type JSONSerializer struct {
	eventTypes map[string]reflect.Type
}

func (j *JSONSerializer) Bind(events ...Event) {
	for _, event := range events {
		eventType, t := EventType(event)
		j.eventTypes[eventType] = t
	}
}

func (j *JSONSerializer) Serialize(v Event) (Record, error) {
	eventType, _ := EventType(v)

	data, err := json.Marshal(v)
	if err != nil {
		return Record{}, err
	}

	data, err = json.Marshal(jsonEvent{
		Type: eventType,
		Data: json.RawMessage(data),
	})
	if err != nil {
		return Record{}, NewError(err, InvalidEncoding, "unable to encode event")
	}

	return Record{
		Version: v.EventVersion(),
		At:      Time(v.EventAt()),
		Data:    data,
	}, nil
}

func (j *JSONSerializer) Deserialize(record Record) (Event, error) {
	wrapper := jsonEvent{}
	err := json.Unmarshal(record.Data, &wrapper)
	if err != nil {
		return nil, NewError(err, InvalidEncoding, "unable to unmarshal event")
	}

	t, ok := j.eventTypes[wrapper.Type]
	if !ok {
		return nil, NewError(err, UnboundEventType, "unbound event type, %v", wrapper.Type)
	}

	v := reflect.New(t).Interface()
	err = json.Unmarshal(wrapper.Data, v)
	if err != nil {
		return nil, NewError(err, InvalidEncoding, "unable to unmarshal event data into %#v", v)
	}

	return v.(Event), nil
}

func NewJSONSerializer(events ...Event) *JSONSerializer {
	serializer := &JSONSerializer{
		eventTypes: map[string]reflect.Type{},
	}
	serializer.Bind(events...)

	return serializer
}
