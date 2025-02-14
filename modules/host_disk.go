// LINUX(HostDiskModule) ok
// WINDOWS(HostDiskModule) ok
// MACOS(HostDiskModule) ?
// ROOT(HostDiskModule) no
package modules

import (
	"fmt"

	"github.com/jaypipes/ghw"
	"github.com/situation-sh/situation/models"
	"github.com/situation-sh/situation/store"
)

func init() {
	RegisterModule(&HostDiskModule{})
}

// Module definition ---------------------------------------------------------

// HostDiskModule retrieves basic information about disk:
// name, model, size, type, controller and partitions.
//
// It heavily relies on the [ghw] library.
//
// On Windows, it uses WMI requests:
//
//		```ps1
//	 SELECT Caption, CreationClassName, Description, DeviceID, FileSystem, FreeSpace, Name, Size, SystemName FROM Win32_LogicalDisk
//	 ```
//
//		```ps1
//		SELECT DeviceId, MediaType FROM MSFT_PhysicalDisk
//	 ```
//
//		```ps1
//		SELECT Access, BlockSize, Caption, CreationClassName, Description, DeviceID, DiskIndex, Index, Name, Size, SystemName, Type FROM Win32_DiskPartition
//	 ```
//
//		```ps1
//		SELECT Antecedent, Dependent FROM Win32_LogicalDiskToPartition
//	 ```
//
// On Linux, it reads `/sys/block/$DEVICE/**` files.
// On windows 11, the local user account must have administrator permissions (it does not mean it must be run as root).
// [ghw]: https://github.com/jaypipes/ghw/
type HostDiskModule struct{}

func (m *HostDiskModule) Name() string {
	return "host-disk"
}

func (m *HostDiskModule) Dependencies() []string {
	return []string{"host-basic"}
}

// see https://pkg.go.dev/github.com/jaypipes/ghw@v0.9.0/pkg/block#StorageController
func diskType(t ghw.DriveType) string {
	switch t {
	case ghw.DRIVE_TYPE_FDD:
		return "floppy"
	case ghw.DRIVE_TYPE_HDD:
		return "hdd"
	case ghw.DRIVE_TYPE_ODD:
		return "optical"
	case ghw.DRIVE_TYPE_SSD:
		return "ssd"
	default:
		return "unknown"
	}
}

// see https://pkg.go.dev/github.com/jaypipes/ghw@v0.9.0/pkg/block#StorageController
func controllerType(t ghw.StorageController) string {
	switch t {
	case ghw.STORAGE_CONTROLLER_IDE:
		// Integrated Drive Electronics
		return "ide"
	case ghw.STORAGE_CONTROLLER_MMC:
		// Multi-media controller (used for mobile phone storage devices)
		return "mmc"
	case ghw.STORAGE_CONTROLLER_NVME:
		// Non-volatile Memory Express
		return "nvme"
	case ghw.STORAGE_CONTROLLER_SCSI:
		// Small computer system interface
		return "scsi"
	case ghw.STORAGE_CONTROLLER_VIRTIO:
		// Virtualized storage controller/driver
		return "virtio"
	default:
		return "unknown"
	}
}

func (m *HostDiskModule) Run() error {
	logger := GetLogger(m)
	machine := store.GetHost()
	if machine == nil {
		return fmt.Errorf("cannot retrieve host machine")
	}

	block, err := ghw.Block(ghw.WithDisableWarnings())
	if err != nil {
		return fmt.Errorf("error while retrieving disk information: %v", err)
	}

	for _, disk := range block.Disks {
		// ignore some disks
		if disk.StorageController == ghw.STORAGE_CONTROLLER_UNKNOWN {
			continue
		}

		// create a disk
		d := models.Disk{
			Name:       disk.Name,
			Model:      disk.Model,
			Size:       disk.SizeBytes,
			Type:       diskType(disk.DriveType),
			Controller: controllerType(disk.StorageController),
			Partitions: make([]*models.Partition, 0),
		}
		logger.WithField("name", d.Name).
			WithField("model", d.Model).
			WithField("size (MiB)", d.Size/(1024*1024)).
			WithField("type", d.Type).
			WithField("controller", d.Controller).
			WithField("#partitions", len(disk.Partitions)).
			Info("Found disk on host")

		// embed partitions
		for _, part := range disk.Partitions {
			p := models.Partition{
				Name:     part.Name,
				Size:     part.SizeBytes,
				Type:     part.Type,
				ReadOnly: part.IsReadOnly,
			}
			d.Partitions = append(d.Partitions, &p)
			logger.WithField("name", p.Name).
				WithField("size (MiB)", p.Size/(1024*1024)).
				WithField("type", p.Type).
				WithField("read_only", p.ReadOnly).
				Info("Here is a partition")
		}

		machine.Disks = append(machine.Disks, &d)
	}
	return nil
}
