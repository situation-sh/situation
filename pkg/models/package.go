package models

import (
	"time"

	"github.com/uptrace/bun"
)

// Package is a wrapper around application that stores distribution
// information of applications (executables)
type Package struct {
	bun.BaseModel `bun:"table:packages"`

	ID        int64     `bun:"id,pk,autoincrement"`
	CreatedAt time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp"`
	UpdatedAt time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp"`

	Name            string `bun:"name,unique:name_version_machine_id" json:"name,omitempty" jsonschema:"description=name of the package,example=openssh,example=musl-gcc,example=python,example=texlive"`
	Version         string `bun:"version,unique:name_version_machine_id" json:"version,omitempty" jsonschema:"description=version of the package,example=8.8p1,example=1.2.3,example=3.10.8,example=2021"`
	Vendor          string `bun:"vendor" json:"vendor,omitempty" jsonschema:"description=name of the organization that produce the package or its maintainer,example=Fedora Project,example=bob@debian.org"`
	Manager         string `bun:"manager" json:"manager,omitempty" jsonschema:"description=program that manages the package installation,example=rpm,example=dpkg,example=msi,example=builtin"`
	InstallTimeUnix int64  `bun:"install_time_unix" json:"install_time,omitempty" jsonschema:"description=UNIX timestamp of the package installation,example=1670520587"`

	Files []string `bun:"files" json:"files"`

	// Belongs-to relationship
	MachineID int64    `bun:"machine_id,notnull,nullzero,unique:name_version_machine_id"`
	Machine   *Machine `bun:"rel:belongs-to,join:machine_id=id,on_delete:cascade"`

	// Has-many relationship
	Applications []*Application `bun:"rel:has-many,join:id=package_id" json:"applications" jsonschema:"description=list of applications"`
}
