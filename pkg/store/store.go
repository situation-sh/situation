package store

import (
	"net"

	"github.com/situation-sh/situation/pkg/models"
)

type Store interface {
	Open() error
	Close() error
	Clear()
	IsHost(*models.Machine) bool
	GetHost() *models.Machine
	SetHost(*models.Machine)
	GetMachineByNetwork(ip net.IP, mac net.HardwareAddr) *models.Machine
	GetMachineByMAC(mac net.HardwareAddr) *models.Machine
	GetMachineByIP(ip net.IP) *models.Machine
	GetMachineByHostID(id string) *models.Machine
	GetMachinesByOpenTCPPort(port uint16) ([]*models.Machine, []*models.Application, []*models.ApplicationEndpoint)
	InsertMachine(m *models.Machine)
	IterateMachines() chan *models.Machine
	GetAllIPv4Networks() []*net.IPNet
	InitPayload() *models.Payload
}
