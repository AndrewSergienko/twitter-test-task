package internal

import "sync"

type EventManager struct {
	targets sync.Map
}

func NewEventManager() *EventManager {
	return &EventManager{targets: sync.Map{}}
}

func (manager *EventManager) SendNewMessageEvent(message Message) {
	manager.targets.Range(func(target any, _ any) bool {
		target.(chan Message) <- message
		return true
	})
}

func (manager *EventManager) AddTarget(target chan Message) {
	manager.targets.Store(target, true)
}

func (manager *EventManager) DeleteTarget(target chan Message) {
	manager.targets.Delete(target)
}
