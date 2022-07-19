package backends

import (
	"context"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/uuid"
	"github.com/situation-sh/situation/models"
)

// GenericTestBackend is a basic function to test a backend
func GenericTestBackend(b Backend, payload *models.Payload) error {
	// fake an http server
	if b.Name() == "http" {
		defaultHttpBackend.url = "http://127.0.0.1:38080/api/discovery/situation/"
		wg := sync.WaitGroup{}
		wg.Add(1)
		srv := runServer(&wg)
		defer func() {
			if err := srv.Shutdown(context.TODO()); err != nil {
				fmt.Println(err)
			}
			wg.Wait()
		}()
	}
	if err := b.Init(); err != nil {
		return err
	}

	b.Write(payload)
	b.Close()
	return nil
}

func TestBackends(t *testing.T) {
	p := randomPayload()
	for name, b := range backends {
		fmt.Printf("--- BACKEND: %s\n", name)
		if err := GenericTestBackend(b, p); err != nil {
			t.Errorf("error with backend %s: %v", name, err)
		}
	}
}

func randomPerformance() models.Performance {
	return models.Performance{
		HeapAlloc: gofakeit.Uint64(),
		HeapSys:   gofakeit.Uint64(),
	}
}

func randomModuleError() *models.ModuleError {
	return &models.ModuleError{
		Module:  gofakeit.Adjective(),
		Message: gofakeit.Sentence(int(gofakeit.Int8())),
	}
}

func randomExtraInfo() *models.ExtraInfo {
	u, _ := uuid.NewRandom()
	return &models.ExtraInfo{
		Agent:     u,
		Version:   gofakeit.AppVersion(),
		Duration:  time.Duration(gofakeit.Uint64()),
		Timestamp: gofakeit.Date(),
		Errors:    []*models.ModuleError{randomModuleError(), randomModuleError()},
		Perfs:     randomPerformance(),
	}
}

func randomApplicationEndpoint() *models.ApplicationEndpoint {
	return &models.ApplicationEndpoint{
		Port:     gofakeit.Uint16(),
		Protocol: gofakeit.RandomString([]string{"tcp", "udp"}),
		Addr:     net.ParseIP(gofakeit.IPv4Address()),
	}
}

func randomApplication() *models.Application {
	return &models.Application{
		Name:      gofakeit.Name(),
		Version:   gofakeit.AppVersion(),
		Endpoints: []*models.ApplicationEndpoint{randomApplicationEndpoint(), randomApplicationEndpoint()},
	}
}

func randomNIC() *models.NetworkInterface {
	mac, _ := net.ParseMAC(gofakeit.MacAddress())
	return &models.NetworkInterface{
		Name:      gofakeit.LetterN(3) + gofakeit.Digit(),
		MAC:       mac,
		IP:        net.ParseIP(gofakeit.IPv4Address()),
		MaskSize:  24,
		IP6:       net.ParseIP(gofakeit.IPv6Address()),
		Mask6Size: 64,
	}
}

func randomCPU() *models.CPU {
	return &models.CPU{
		ModelName: gofakeit.NounAbstract(),
		Vendor:    gofakeit.Company(),
		Cores:     gofakeit.IntRange(4, 16),
	}
}

func randomMachine() *models.Machine {
	return &models.Machine{
		Hostname:            gofakeit.NounCommon(),
		HostID:              gofakeit.UUID(),
		Arch:                gofakeit.RandomString([]string{"amd64", "386", "aarch64"}),
		Platform:            gofakeit.RandomString([]string{"windows", "linux"}),
		Distribution:        gofakeit.NounCommon(),
		DistributionVersion: gofakeit.AppVersion(),
		ParentMachine:       nil,
		CPU:                 randomCPU(),
		NICS:                []*models.NetworkInterface{randomNIC(), randomNIC()},
		Applications: []*models.Application{
			randomApplication(),
			randomApplication(),
			randomApplication(),
			randomApplication()},
		Uptime: time.Duration(gofakeit.Uint64()),
	}
}

func randomPayload() *models.Payload {
	return &models.Payload{
		Machines: []*models.Machine{randomMachine(), randomMachine(), randomMachine()},
		Extra:    randomExtraInfo(),
	}
}
