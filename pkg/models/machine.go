package models

import (
	"time"

	"github.com/uptrace/bun"
)

var machineCounter = 0

// Machine is a generic structure to represent nodes on
// an information system. It can be a physical machine,
// a VM, a container...
type Machine struct {
	bun.BaseModel `bun:"table:machines,alias:machine"`

	ID        int64     `bun:"id,pk,autoincrement"`
	CreatedAt time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp"`
	UpdatedAt time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp"`

	Hostname            string        `bun:"hostname" json:"hostname,omitempty" jsonschema:"description=name of the machine,example=DESKTOP-2HHPC7I,example=PC-JEAN-LUC,example=server07"`
	HostID              string        `bun:"host_id,unique,nullzero" json:"host_id,omitempty" jsonschema:"description=machine uuid identifier,example=8375c6c3-de33-41a4-bdb2-4e467d9f632c"`
	Arch                string        `bun:"arch" json:"arch,omitempty" jsonschema:"description=architecture,example=x86_64"`
	Platform            string        `bun:"platform" json:"platform,omitempty" jsonschema:"description=system base platform,example=linux,example=windows,example=docker"`
	Distribution        string        `bun:"distribution" json:"distribution,omitempty" jsonschema:"description=OS name (or base image),example=fedora,example=Microsoft Windows 10 Home,example=postgres"`
	DistributionVersion string        `bun:"distribution_version" json:"distribution_version,omitempty" jsonschema:"description=OS version (or image version),example=36,example=10.0.19044 Build 19044,example=latest"`
	DistributionFamily  string        `bun:"distribution_family" json:"distribution_family,omitempty" jsonschema:"description=OS family,example=debian,example=fedora,example=rhel,example=suse,example=Standalone Workstation,example=Server,example=Server (Domain Controller)"`
	Uptime              time.Duration `bun:"uptime" json:"uptime,omitempty" jsonschema:"description=machine uptime in nanoseconds,example=22343000000000,example=13521178203519"`

	Agent string `bun:"agent,unique,nullzero" json:"agent,omitempty" jsonschema:"description=collector identifier (agent case)"`

	CPE     string `json:"cpe,omitempty" jsonschema:"description=OS CPE uri,example=cpe:2.3:o:microsoft:windows_server_2022:-:*:*:*:datacenter:*:x64:*"`
	Chassis string `json:"chassis,omitempty" jsonschema:"description=machine kind,example=vm,example=laptop"`

	// Has-one relationship
	ParentMachineID int64    `bun:"parent_machine_id,nullzero" json:"parent_machine,omitempty" jsonschema:"description=internal reference of the parent machine (docker or VM cases especially),example=53127"`
	ParentMachine   *Machine `bun:"rel:has-one,join:parent_machine_id=id" json:"parent,omitempty" jsonschema:"description=parent machine (docker or VM host)"`

	// Has-one relationship
	CPU *CPU `bun:"rel:has-one,join:id=machine_id" json:"cpu,omitempty" jsonschema:"description=CPU information"`

	// Has-many relationship
	Packages []*Package `bun:"rel:has-many,join:id=machine_id" json:"packages" jsonschema:"description=list of packages"`

	// Has-many relationship
	NICS []*NetworkInterface `bun:"rel:has-many,join:id=machine_id" json:"nics" jsonschema:"description=list of network devices"`

	// Has-many relationship
	Disks []*Disk `bun:"rel:has-many,join:id=machine_id" json:"disks" jsonschema:"description=list of disks"`

	// Has-many relationship
	GPUS []*GPU `bun:"rel:has-many,join:id=machine_id" json:"gpus" jsonschema:"description=list of GPU"`

	// Has-many relationship
	Applications []*Application `bun:"rel:has-many,join:id=machine_id" json:"applications" jsonschema:"description=list of applications"`
}

// NewMachine inits a new Machine structure
func NewMachine() *Machine {
	// increment ID
	machineCounter++
	// return new machine
	return &Machine{
		// InternalID: machineCounter, // ensure the InternalId is greater than 1
		Packages: make([]*Package, 0),
		NICS:     make([]*NetworkInterface, 0),
		Disks:    make([]*Disk, 0),
		GPUS:     make([]*GPU, 0),
	}
}

func (m *Machine) GetPackageByApplicationPath(path string) *Package {
	for _, p := range m.Packages {
		for _, a := range p.Applications {
			if a.Name == path {
				return p
			}
		}
	}
	return nil
}
