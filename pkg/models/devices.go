package models

import (
	"time"

	"github.com/uptrace/bun"
)

// CPU gathers few information about processor
type CPU struct {
	bun.BaseModel `bun:"table:cpus"`

	ID        int64     `bun:"id,pk,autoincrement"`
	CreatedAt time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp"`
	UpdatedAt time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp"`

	ModelName string `bun:"model_name" json:"model_name,omitempty" jsonschema:"description=common name of the CPU,example=Intel Xeon Processor (Cooperlake),example=Intel(R) Core(TM) i7-10850H CPU @ 2.70GHz"`
	Vendor    string `bun:"vendor" json:"vendor,omitempty" jsonschema:"description=CPU 12-chars vendor string,example=GenuineIntel,example=AuthenticAMD,minLength=12,maxLength=12"`
	Cores     int    `bun:"cores" json:"cores,omitempty" jsonschema:"description=number of cores,example=2,example=6"`

	// Belongs-to relationship
	MachineID int64    `bun:"machine_id,notnull"`
	Machine   *Machine `bun:"rel:belongs-to,join:machine_id=id"`
}

type Disk struct {
	bun.BaseModel `bun:"table:disks"`

	ID        int64     `bun:"id,pk,autoincrement"`
	CreatedAt time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp"`
	UpdatedAt time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp"`

	Name       string      `bun:"name,unique:machine_disk_name" json:"name,omitempty" jsonschema:"description=name of the device,example=nvme0n1,example=\\\\.\\PHYSICALDRIVE0"`
	Model      string      `bun:"model" json:"model,omitempty" jsonschema:"description=model the device,example=Micron 2300 NVMe 512GB,example=INTEL SSDSC2BF240A5L"`
	Size       uint64      `bun:"size" json:"size,omitempty" jsonschema:"description=disk size in bytes,example=512110190592,example=240054796800"`
	Type       string      `bun:"type" json:"type,omitempty" jsonschema:"description=disk type,example=ssd,example=hdd,example=optical,example=floppy"`
	Controller string      `bun:"controller" json:"controller,omitempty" jsonschema:"description=physical hardware interface,example=ide,example=scsi,example=nvme,example=virtio"`
	Partitions []Partition `bun:"partitions,type:json" json:"partitions,omitempty" jsonschema:"description=list of partitions"`

	// Belongs-to relationship
	MachineID int64    `bun:"machine_id,unique:machine_disk_name,notnull"`
	Machine   *Machine `bun:"rel:belongs-to,join:machine_id=id"`
}

type Partition struct {
	Name     string `json:"name,omitempty" jsonschema:"description=name of the partition,example=nvme0n1p1"`
	Size     uint64 `json:"size,omitempty" jsonschema:"description=size of the partition in bytes,example=629145600"`
	Type     string `json:"type,omitempty" jsonschema:"description=filesystem,example=vfat,example=ext4,example=crypto_LUKS"`
	ReadOnly bool   `json:"read_only,omitempty" jsonschema:"description=if the partition is writeable"`
}

type GPU struct {
	bun.BaseModel `bun:"table:gpus"`

	ID        int64     `bun:"id,pk,autoincrement"`
	CreatedAt time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp"`
	UpdatedAt time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp"`

	Index   int    `bun:"index,unique:machine_gpu_index" json:"index" jsonschema:"description=index of the device (relevant when there are several),example=0,example=1"`
	Product string `bun:"product" json:"product,omitempty" jsonschema:"description=product name,example=CometLake-H GT2 [UHD Graphics],example=Intel(R) HD Graphics 530,example=TU117GLM [Quadro T2000 Mobile / Max-Q]"`
	Vendor  string `bun:"vendor" json:"vendor,omitempty" jsonschema:"description=manufacturer name,example=Intel Corporation,example=NVIDIA Corporation"`
	Driver  string `bun:"driver" json:"driver,omitempty" jsonschema:"description=GPU driver name,example=i915,example=nouveau"`

	// Belongs-to relationship
	MachineID int64    `bun:"machine_id,unique:machine_gpu_index,notnull"`
	Machine   *Machine `bun:"rel:belongs-to,join:machine_id=id"`
}
