package models

import (
	"time"

	"github.com/google/uuid"
)

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
	HeapAlloc uint64 `json:"heap_alloc" jsonschema:"description=bytes allocated in the heap that represent reachable objects"`

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
	HeapSys uint64 `json:"heap_sys" jsonschema:"description=the amount of virtual address space reserved for the heap (it estimates the largest size the heap has had)"`
}

// ModuleError is a structure to send possible errors
// at the module level
type ModuleError struct {
	Module  string `json:"module" jsonschema:"description=name of the module,example=docker,example=rpm"`
	Message string `json:"message" jsonschema:"description=error message"`
}

// ExtraInfo stores extra agent/scan informations
type ExtraInfo struct {
	Agent     uuid.UUID      `json:"agent" jsonschema:"description=agent uuid identifier,example=cafecafe-cafe-cafe-cafe-cafecafecafe"`
	Version   string         `json:"version" jsonschema:"description=agent version,example=0.13.2"`
	Duration  time.Duration  `json:"duration" jsonschema:"description=scan duration in nanoseconds,example=2010899300"`
	Timestamp time.Time      `json:"timestamp" jsonschema:"description=timestamp of the end of the scan,example=2022-12-09T10:47:34.0210722+01:00"`
	Errors    []*ModuleError `json:"errors" jsonschema:"description=list of the encountered errors"`
	Perfs     Performance    `json:"perfs" jsonschema:"description=agent performances"`
}

// Payload is the full data that is sent to the server
type Payload struct {
	Machines []*Machine `json:"machines" jsonschema:"description=list of the machines"`
	Extra    *ExtraInfo `json:"extra" jsonschema:"description=scan extra information (agent only)"`
}
