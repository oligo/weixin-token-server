package main

import "os"

type Hook struct {
	sigChan chan os.Signal

	handlers []HookCallback
}

// HookCallback declares the API for the registered handlers
type HookCallback func()

func (h *Hook) register(c HookCallback) {
	h.handlers = append(h.handlers, c)
}

func (h *Hook) listen() {
	for {
		select {
		case <-signalChan:
			for _, handler := range h.handlers {
				handler()
			}
		default:

		}
	}
}
