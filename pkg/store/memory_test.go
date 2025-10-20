package store

import (
	"testing"

	"github.com/google/uuid"
)

func TestMemoryStore(t *testing.T) {
	agent := uuid.New()
	gst := &GenericStoreTester{
		s: NewMemoryStore(agent),
		t: t,
	}
	gst.TestClear()
	gst.TestGetHost()
	gst.TestGetMachineByNetwork()
	gst.TestGetAllIPv4Networks()
	gst.TestGetMachineByHostID()
	gst.TestGetMachinesByOpenTCPPort()
	gst.TestIterateMachines()
}
