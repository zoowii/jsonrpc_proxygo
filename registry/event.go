package registry

import "fmt"

type EventType string

const (
	SERVICE_ADD    EventType = "service_add"
	SERVICE_REMOVE EventType = "service_remove"
)

type Event struct {
	Type        EventType
	ServiceInfo *Service
}

func NewEvent(eventType EventType, serviceInfo *Service) *Event {
	return &Event{
		Type:        eventType,
		ServiceInfo: serviceInfo,
	}
}

func (e *Event) String() string {
	return fmt.Sprintf("Event<Type=%s, ServiceInfo=%s>", e.Type, e.ServiceInfo.String())
}
