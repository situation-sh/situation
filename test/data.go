package test

import (
	"net"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/uuid"
	"github.com/situation-sh/situation/models"
)

func RandomPerformance() models.Performance {
	return models.Performance{
		HeapAlloc: gofakeit.Uint64(),
		HeapSys:   gofakeit.Uint64(),
	}
}

func RandomModuleError() *models.ModuleError {
	return &models.ModuleError{
		Module:  gofakeit.Adjective(),
		Message: gofakeit.Sentence(int(gofakeit.Int8())),
	}
}

func RandomExtraInfo() *models.ExtraInfo {
	u, _ := uuid.NewRandom()
	return &models.ExtraInfo{
		Agent:     u,
		Version:   gofakeit.AppVersion(),
		Duration:  time.Duration(gofakeit.Int64()),
		Timestamp: gofakeit.Date(),
		Errors:    []*models.ModuleError{RandomModuleError(), RandomModuleError()},
		Perfs:     RandomPerformance(),
	}
}

func RandomApplicationEndpoint() *models.ApplicationEndpoint {
	return &models.ApplicationEndpoint{
		Port:     gofakeit.Uint16(),
		Protocol: gofakeit.RandomString([]string{"tcp", "udp"}),
		Addr:     net.ParseIP(gofakeit.IPv4Address()),
	}
}

func RandomApplication() *models.Application {
	return &models.Application{
		Name:      gofakeit.Name(),
		Args:      []string{gofakeit.Name(), gofakeit.Name(), gofakeit.Name()},
		Endpoints: []*models.ApplicationEndpoint{RandomApplicationEndpoint(), RandomApplicationEndpoint()},
	}
}

func RandomPackage() *models.Package {
	return &models.Package{
		Name:            gofakeit.Name(),
		Version:         gofakeit.AppVersion(),
		Vendor:          gofakeit.Company(),
		Manager:         gofakeit.RandomString([]string{"msi", "builtin", "manual", "rpm", "dpkg"}),
		InstallTimeUnix: gofakeit.Date().Unix(),
		Applications:    []*models.Application{RandomApplication(), RandomApplication()},
	}
}

func RandomNICFlags() *models.NetworkInterfaceFlags {
	return &models.NetworkInterfaceFlags{
		Up:           gofakeit.Bool(),
		Broadcast:    gofakeit.Bool(),
		Loopback:     gofakeit.Bool(),
		PointToPoint: gofakeit.Bool(),
		Multicast:    gofakeit.Bool(),
		Running:      gofakeit.Bool(),
	}
}

func RandomNIC() *models.NetworkInterface {
	mac, _ := net.ParseMAC(gofakeit.MacAddress())
	return &models.NetworkInterface{
		Name:      gofakeit.LetterN(3) + gofakeit.Digit(),
		MAC:       mac,
		IP:        net.ParseIP(gofakeit.IPv4Address()),
		MaskSize:  24,
		IP6:       net.ParseIP(gofakeit.IPv6Address()),
		Mask6Size: 64,
		Gateway:   net.ParseIP(gofakeit.IPv4Address()),
		Flags:     RandomNICFlags(),
	}
}

func RandomCPU() *models.CPU {
	return &models.CPU{
		ModelName: gofakeit.NounAbstract(),
		Vendor:    gofakeit.Company(),
		Cores:     gofakeit.IntRange(4, 16),
	}
}

func RandomPartition() *models.Partition {
	return &models.Partition{
		Name:     gofakeit.LetterN(3) + gofakeit.Digit(),
		Size:     gofakeit.Uint64(),
		Type:     gofakeit.LetterN(5),
		ReadOnly: gofakeit.Bool(),
	}
}

func RandomDisk() *models.Disk {
	n := gofakeit.IntRange(1, 5)
	partitions := make([]*models.Partition, n)
	for i := 0; i < n; i++ {
		partitions[i] = RandomPartition()
	}

	return &models.Disk{
		Name:       gofakeit.LetterN(3),
		Model:      gofakeit.Company() + "_" + gofakeit.UUID(),
		Size:       gofakeit.Uint64(),
		Type:       gofakeit.RandomString([]string{"floppy", "hdd", "optical", "ssd", "unknown"}),
		Controller: gofakeit.RandomString([]string{"ide", "mmc", "nvme", "scsi", "virtio", "unknown"}),
		Partitions: partitions,
	}
}

func RandomGPU() *models.GPU {
	return &models.GPU{
		Index:   gofakeit.IntRange(0, 5),
		Vendor:  gofakeit.Company(),
		Product: gofakeit.NounAbstract(),
		Driver:  gofakeit.FarmAnimal(),
	}
}

func RandomMachine() *models.Machine {
	return &models.Machine{
		InternalID:          gofakeit.IntRange(1, 10000),
		Hostname:            gofakeit.NounCommon(),
		HostID:              gofakeit.UUID(),
		Arch:                gofakeit.RandomString([]string{"amd64", "386", "aarch64"}),
		Platform:            gofakeit.RandomString([]string{"windows", "linux"}),
		Distribution:        gofakeit.NounCommon(),
		DistributionVersion: gofakeit.AppVersion(),
		ParentMachine:       -1,
		CPU:                 RandomCPU(),
		NICS:                []*models.NetworkInterface{RandomNIC(), RandomNIC()},
		Disks:               []*models.Disk{RandomDisk()},
		GPUS:                []*models.GPU{RandomGPU()},
		Packages: []*models.Package{
			RandomPackage(),
			RandomPackage(),
			RandomPackage(),
			RandomPackage()},
		Uptime: time.Duration(gofakeit.Int64()),
	}
}

func RandomHostMachine() *models.Machine {
	m := RandomMachine()
	u, err := uuid.Parse(gofakeit.UUID())
	if err == nil {
		m.Agent = &u
	}
	return m
}

func RandomPayload() *models.Payload {
	return &models.Payload{
		Machines: []*models.Machine{RandomMachine(), RandomMachine(), RandomMachine()},
		Extra:    RandomExtraInfo(),
	}
}
