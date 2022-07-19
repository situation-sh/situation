package models

import (
	"time"

	"github.com/google/uuid"
)

// CPU gathers few information about processor
type CPU struct {
	ModelName string `json:"model_name,omitempty"`
	Vendor    string `json:"vendor,omitempty"`
	Cores     int    `json:"cores,omitempty"`
	// Usage     float64 `json:"usage"`
}

// Performance gathers information about the run of
// the agent
type Performance struct {
	// HeapAlloc is bytes of allocated heap objects.
	//
	// "Allocated" heap objects include all reachable objects, as
	// well as unreachable objects that the garbage collector has
	// not yet freed. Specifically, HeapAlloc increases as heap
	// objects are allocated and decreases as the heap is swept
	// and unreachable objects are freed. Sweeping occurs
	// incrementally between GC cycles, so these two processes
	// occur simultaneously, and as a result HeapAlloc tends to
	// change smoothly (in contrast with the sawtooth that is
	// typical of stop-the-world garbage collectors).
	HeapAlloc uint64 `json:"heap_alloc"`

	// HeapSys is bytes of heap memory obtained from the OS.
	//
	// HeapSys measures the amount of virtual address space
	// reserved for the heap. This includes virtual address space
	// that has been reserved but not yet used, which consumes no
	// physical memory, but tends to be small, as well as virtual
	// address space for which the physical memory has been
	// returned to the OS after it became unused (see HeapReleased
	// for a measure of the latter).
	//
	// HeapSys estimates the largest size the heap has had.
	HeapSys uint64 `json:"heap_sys"`
}

// ModuleError is a structure to send possible errors
// at the module level
type ModuleError struct {
	Module  string `json:"module"`
	Message string `json:"message"`
}

// ExtraInfo stores extra agent/scan informations
type ExtraInfo struct {
	Agent     uuid.UUID      `json:"agent"`
	Version   string         `json:"version"`
	Duration  time.Duration  `json:"duration"`
	Timestamp time.Time      `json:"timestamp"`
	Errors    []*ModuleError `json:"errors"`
	Perfs     Performance    `json:"perfs"`
}

// Payload is the full data that is sent to the server
type Payload struct {
	Machines []*Machine `json:"machines"`
	Extra    *ExtraInfo `json:"extra"`
}
