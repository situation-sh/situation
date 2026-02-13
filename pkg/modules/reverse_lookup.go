// LINUX(ReverseLookupModule) ok
// WINDOWS(ReverseLookupModule) ok
// MACOS(ReverseLookupModule) ?
// ROOT(ReverseLookupModule) ?
package modules

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/situation-sh/situation/pkg/models"
)

func init() {
	registerModule(&ReverseLookupModule{})
}

// ReverseLookupModule tries to get a hostname attached to a local IP address
//
// It basically calls [net.LookupAddr]
// that uses the host resolver to perform a reverse lookup for the given addresses.
//
// [net.LookupAddr]: https://pkg.go.dev/net#LookupAddr
type ReverseLookupModule struct {
	BaseModule
}

func (m *ReverseLookupModule) Name() string {
	return "reverse-lookup"
}

func (m *ReverseLookupModule) Dependencies() []string {
	// depends on ping to ensure a rather fresh
	// arp table
	return []string{"arp"}
}

func (m *ReverseLookupModule) Run(ctx context.Context) error {
	logger := getLogger(ctx, m)
	storage := getStorage(ctx)

	nics := make([]*models.NetworkInterface, 0)
	err := storage.DB().
		NewSelect().
		Model(&nics).
		Where("machine_id IS NULL").
		Scan(ctx)
	if err != nil {
		return fmt.Errorf("failed to retrieve orphan nics: %v", err)
	}

	if len(nics) == 0 {
		logger.Info("No orphan NICs found, skipping reverse lookup")
		return nil
	}

	newMachines := make([]*models.Machine, 0)
	for _, nic := range nics {
		for _, ip := range nic.IPs() {
			if ip == nil || ip.IsLoopback() {
				continue
			}
			// if ip != nil && ip.IsPrivate() {
			// run first lookup
			net.LookupAddr(ip.String()) // #nosec G104 -- we don't care about the errors here
			names, err := net.LookupAddr(ip.String())
			if err != nil {
				logger.
					WithField("ip", ip).
					WithError(err).
					Warn("Reverse lookup failed")
				continue
			}
			if len(names) > 0 {
				m := &models.Machine{
					Hostname: strings.TrimSuffix(names[0], "."),
				}
				newMachines = append(newMachines, m)
				nic.Machine = m
				logger.
					WithField("hostname", m.Hostname).
					WithField("ip", ip).
					Info("Hostname resolved")
			} else {
				logger.
					WithField("ip", ip).
					Info("No hostname found for IP")
			}
		}
	}

	if len(newMachines) > 0 {
		_, err = storage.DB().
			NewInsert().
			Model(&newMachines).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to insert new machines: %v", err)
		}
		logger.
			WithField("machines", len(newMachines)).
			Info("Inserted new machines with hostnames")

		// upate nics with machine_id
		nicsToUpdate := make([]*models.NetworkInterface, 0)
		for _, nic := range nics {
			if nic.Machine != nil {
				nic.MachineID = nic.Machine.ID
				nicsToUpdate = append(nicsToUpdate, nic)
			}
		}

		if len(nicsToUpdate) > 0 {
			_, err = storage.DB().
				NewUpdate().
				Model(&nicsToUpdate).
				Column("machine_id").
				Bulk().
				Exec(ctx)
			if err != nil {
				return fmt.Errorf("failed to update nics: %v", err)
			}
			logger.WithField("nics", len(nicsToUpdate)).
				Info("Updated nics with machine_id")
		}
	}

	// machinesWithoutHostname := make([]*models.Machine, 0)
	// err := storage.DB().
	// 	NewSelect().
	// 	Model(&machinesWithoutHostname).
	// 	Where("hostname = '' OR hostname IS NULL").
	// 	Relation("NICS").
	// 	Scan(ctx)
	// if err != nil {
	// 	logger.WithError(err).Error("Cannot retrieve machines without hostname")
	// 	return err
	// }
	// if len(machinesWithoutHostname) == 0 {
	// 	logger.Info("No machine without hostname found")
	// 	return nil
	// }

	// machinesToUpdate := make([]*models.Machine, 0)

	// outer:
	// 	for _, machine := range machinesWithoutHostname {

	// 		for _, nic := range machine.NICS {
	// 			for _, ip := range nic.IPs() {
	// 				if ip != nil && ip.IsPrivate() {
	// 					// run first lookup
	// 					net.LookupAddr(ip.String()) // #nosec G104 -- we don't care about the errors here
	// 					names, err := net.LookupAddr(ip.String())

	// 					if err != nil {
	// 						logger.
	// 							WithField("ip", ip).
	// 							WithError(err).
	// 							Warn("Reverse lookup failed")
	// 						continue
	// 					}
	// 					if len(names) > 0 {
	// 						// on linux we may have an ending dot
	// 						machine.Hostname = strings.TrimSuffix(names[0], ".")
	// 						machinesToUpdate = append(machinesToUpdate, machine)
	// 						logger.
	// 							WithField("hostname", machine.Hostname).
	// 							WithField("ip", ip).
	// 							Info("Hostname resolved")
	// 						// go to the next machine
	// 						continue outer
	// 					}
	// 				}
	// 			}
	// 		}

	// 	}
	// 	if len(machinesToUpdate) > 0 {
	// 		_, err = storage.DB().NewUpdate().
	// 			Model(&machinesToUpdate).
	// 			Column("hostname", "updated_at").
	// 			Bulk().
	// 			Exec(ctx)
	// 		if err != nil {
	// 			logger.WithError(err).Error("Failed to bulk update hostnames")
	// 			return err
	// 		}
	// 		logger.WithField("machines", len(machinesToUpdate)).
	// 			Info("Bulk updated machine hostnames")
	// 	} else {
	// 		logger.Info("No machine to update")
	// 	}

	return nil
}
