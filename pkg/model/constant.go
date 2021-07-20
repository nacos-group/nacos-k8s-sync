package model

import "time"

// Event represents a registry update event
type Event int

type Direction string

const (
	// EventAdd is sent when an object is added
	EventAdd Event = iota

	// EventUpdate is sent when an object is modified
	// Captures the modified object
	EventUpdate

	// EventDelete is sent when an object is deleted
	// Captures the object at the last known state
	EventDelete

	DefaultTaskDelay = 1 * time.Second

	DefaultResyncInterval = 0

	DefaultNacosEndpointWeight = 100

	MaxRetry = 3

	ToNacos Direction = "to-nacos"

	ToK8s Direction = "to-k8s"

	Both Direction = "both"
)
