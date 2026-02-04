// LINUX(NetstatModule) ok
// WINDOWS(NetstatModule) ok
// MACOS(NetstatModule) ?
// ROOT(NetstatModule) yes
package modules

import (
	"context"
	"fmt"
	"os"
	"os/user"
	"runtime"

	"github.com/cakturk/go-netstat/netstat"
	"github.com/situation-sh/situation/pkg/models"
	"github.com/situation-sh/situation/pkg/store"
	"github.com/situation-sh/situation/pkg/utils"
)

var (
	netstatProviders = []netstatProvider{netstat.TCPSocks, netstat.UDPSocks, netstat.TCP6Socks, netstat.UDP6Socks}
	netstatProtocols = []string{"tcp", "udp", "tcp6", "udp6"}
	acceptAll        = func(e *netstat.SockTabEntry) bool { return true }
)

func init() {
	registerModule(&NetstatModule{})
}

// Module definition ---------------------------------------------------------

// NetstatModule retrieves active connections.
//
// It enumerates TCP, UDP, TCP6 and UDP6 sockets to discover listening
// endpoints, running applications (with PID and command line), and
// network flows between them. It must be run as root on Linux to
// retrieve PID/process information; without these data it is hard
// to build reliable links between open ports and programs.
//
// This module is then able to create flows between applications according
// to the tuple (src, srcport, dst, dstport).
//
// On Windows, the privileges are not checked. So the module is always run.
//
// [go-netstat]: https://github.com/cakturk/go-netstat
type NetstatModule struct {
	BaseModule
}

func (m *NetstatModule) Name() string {
	return "netstat"
}

func (m *NetstatModule) Dependencies() []string {
	return []string{"local-users", "tcp-scan"}
}

func flowFilter(state netstat.SkState) bool {
	for _, s := range []netstat.SkState{
		netstat.Established,
		netstat.FinWait1,
		netstat.FinWait2,
		netstat.TimeWait,
		netstat.CloseWait,
		netstat.LastAck,
		netstat.Closing} {
		if s == state {
			return true
		}
	}
	return false
}

// portFilter returns true when the connection is listening, established or close-wait
// func portFilter(e *netstat.SockTabEntry) bool {
// 	// accept everything
// 	if e.State == netstat.Listen || flowFilter(e.State) {
// 		return true
// 	}
// 	// fmt.Println(e.LocalAddr.IP, ":", e.LocalAddr.Port, " -> ", e.RemoteAddr.IP, ":", e.RemoteAddr.Port)
// 	// if e.LocalAddr.IP.IsLoopback() && !e.RemoteAddr.IP.IsLoopback() {
// 	// 	return false
// 	// }

// 	return false
// }

func portFilter(e *netstat.SockTabEntry) bool {
	// accept everything except Listen state
	if flowFilter(e.State) {
		return true
	}
	// fmt.Println(e.LocalAddr.IP, ":", e.LocalAddr.Port, " -> ", e.RemoteAddr.IP, ":", e.RemoteAddr.Port)
	// if e.LocalAddr.IP.IsLoopback() && !e.RemoteAddr.IP.IsLoopback() {
	// 	return false
	// }

	return false
}

// helper for the Run function
type netstatProvider func(accept netstat.AcceptFn) ([]netstat.SockTabEntry, error)

func hashEndpoint(e *models.ApplicationEndpoint) string {
	return fmt.Sprintf("%s:%d/%s", e.Addr, e.Port, e.Protocol)
}

func hashApplication(a *models.Application) string {
	return fmt.Sprintf("%s/%d", a.Name, a.PID)
}

func hashFlow(f *models.Flow) string {
	return fmt.Sprintf("%d/%s/%d", f.SrcApplicationID, f.SrcAddr, f.DstEndpointID)
}

func hashUserApp(ua *models.UserApplication) string {
	return fmt.Sprintf("%d/%s", ua.Application.PID, ua.User.UID)
}

// buildLocalIPToNICsMap builds a map of IP addresses to network interfaces for the local machine.
// It also maps wildcard addresses (0.0.0.0 and ::) to all local NICs.
func buildLocalIPToNICsMap(ctx context.Context, storage *store.BunStorage, machineID int64) map[string][]*models.NetworkInterface {
	ipMapper := map[string][]*models.NetworkInterface{}
	localNICS := storage.GetMachineNICs(ctx, machineID)
	for _, nic := range localNICS {
		for _, ip := range nic.IP {
			ipMapper[ip] = append(ipMapper[ip], nic)
		}
		ipMapper["0.0.0.0"] = append(ipMapper["0.0.0.0"], nic)
		ipMapper["::"] = append(ipMapper["::"], nic)
	}
	return ipMapper
}

func buildIPNICMap(ctx context.Context, storage *store.BunStorage) (map[string]*models.NetworkInterface, error) {
	allIPs := make(map[string]bool)
	for _, provider := range netstatProviders {
		entries, err := provider(acceptAll)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			// grab all the IPs visible in the netstat entries
			for _, ip := range []string{entry.LocalAddr.IP.String(), entry.RemoteAddr.IP.String()} {
				if ip != "" && ip != "0.0.0.0" && ip != "::" {
					allIPs[ip] = true
				}
			}
		}
	}

	ipList := make([]string, 0, len(allIPs))
	for ip := range allIPs {
		ipList = append(ipList, ip)
	}

	nics, err := storage.GetNICsByIPs(ctx, ipList)
	if err != nil {
		return nil, err
	}
	ipNICMap := make(map[string]*models.NetworkInterface)
	for i := range nics {
		nic := &nics[i]
		for _, ip := range nic.IP {
			ipNICMap[ip] = nic
		}
	}
	return ipNICMap, nil
}

func buildUIDUserMap(ctx context.Context, storage *store.BunStorage) (map[string]*models.User, error) {
	users, err := storage.GetLocalUsers(ctx)
	if err != nil {
		return nil, err
	}
	uidUserMap := make(map[string]*models.User)
	for i := range users {
		user := &users[i]
		uidUserMap[user.UID] = user
	}
	return uidUserMap, nil
}

func (m *NetstatModule) Run(ctx context.Context) error {

	logger := getLogger(ctx, m)
	storage := getStorage(ctx)

	u, err := user.Current()
	if err != nil {
		return err
	}

	if runtime.GOOS == "linux" && u.Uid != "0" {
		logger.WithField("uid", u.Uid).Warn("The module must be run as root")
		// logger.Warnf("On Linux, the %s module must be run as root", m.Name())
		return nil
	}

	machine := storage.GetOrCreateHost(ctx)
	if machine == nil {
		return fmt.Errorf("unable to create or retrieve host machine")
	}

	// build a map of IPs to NICs for remote endpoints matching
	ipNICMap, err := buildIPNICMap(ctx, storage)
	if err != nil {
		return fmt.Errorf("fail to build IP -> NIC mapper: %v", err)
	}
	// build a map of local NICs by IP for later use
	ipMapper := buildLocalIPToNICsMap(ctx, storage, machine.ID)

	// build a map of UID to User for app <-> user linking
	uidUserMap, err := buildUIDUserMap(ctx, storage)
	if err != nil {
		return fmt.Errorf("fail to build UID -> User mapper: %v", err)
	}

	// data collectors
	apps := make([]*models.Application, 0)
	uniqueApps := make(map[string]*models.Application)
	endpoints := make([]*models.ApplicationEndpoint, 0)
	uniqueEnpoints := make(map[string]*models.ApplicationEndpoint)
	flows := make([]*models.Flow, 0)
	uniqueFlows := make(map[string]*models.Flow)
	userApps := make([]*models.UserApplication, 0)
	uniqueUserApps := make(map[string]*models.UserApplication)

	// collects all the local listening endpoints to drop some netstat entries in the next step
	listeningAddrs := make(map[string]bool)
	for k, provider := range netstatProviders {
		entries, err := provider(func(e *netstat.SockTabEntry) bool {
			return true
		})
		if err != nil {
			logger.Errorf("Error while retrieving %s connections: %v", netstatProtocols[k], err)
			continue
		}

		for _, entry := range entries {
			if entry.State != netstat.Listen || entry.LocalAddr == nil {
				continue
			}
			// if entry.State == netstat.Listen && entry.LocalAddr != nil {
			ip := entry.LocalAddr.IP.String()
			// add the address in all cases
			listeningAddrs[entry.LocalAddr.String()] = true
			// then look for all nics that match this ip (in case of 0.0.0.0 or ::)
			for _, nic := range ipMapper[ip] {
				if nic == nil {
					continue
				}
				for _, nicIP := range nic.IP {
					// re-build the address (see netstat module)
					key := fmt.Sprintf("%s:%d", nicIP, entry.LocalAddr.Port)
					listeningAddrs[key] = true

					// add app + endpoint
					name := entry.Process.Name
					args, err := utils.GetCmd(entry.Process.Pid)
					if err == nil && len(args) > 0 {
						name = args[0]
						args = args[1:]
					}

					app := &models.Application{Name: name, Args: args, MachineID: machine.ID}
					pid := entry.Process.Pid
					if pid >= 0 {
						// populate PID
						app.PID = uint64(pid)
						// populate user-app link
						u, err := utils.GetProcessUser(pid)
						if err == nil {
							if localUser, exists := uidUserMap[u.Uid]; exists {
								userApp := &models.UserApplication{
									UserID:      localUser.ID,
									User:        localUser,
									Application: app,
								}
								hUserApp := hashUserApp(userApp)
								if _, exists := uniqueUserApps[hUserApp]; !exists {
									uniqueUserApps[hUserApp] = userApp
								}
							}
						}
					}

					hApp := hashApplication(app)
					if existingApp, exists := uniqueApps[hApp]; exists {
						app = existingApp
					} else {
						uniqueApps[hApp] = app
					}

					endpoint := &models.ApplicationEndpoint{
						Addr:               nicIP,
						Port:               entry.LocalAddr.Port,
						Protocol:           netstatProtocols[k],
						NetworkInterfaceID: nic.ID,
						Application:        app,
					}
					h := hashEndpoint(endpoint)
					if _, exists := uniqueEnpoints[h]; !exists {
						uniqueEnpoints[h] = endpoint
					}
				}
			}
		}
	}

	isListening := func(addr string) bool {
		listens, exists := listeningAddrs[addr]
		return exists && listens
	}

	// loop over all providers
	for k, provider := range netstatProviders {
		// list all entries by protocol
		if entries, err := provider(portFilter); err == nil {
			// here we do not have Listen connection (only true flows)
			for _, entry := range entries {
				if entry.Process == nil {
					continue
				}
				if entry.Process.Pid == os.Getpid() {
					// ignore self
					continue
				}
				// ignore docker-proxy
				// this process aims at forwarding port
				// if entry.Process.Name == "docker-proxy" {
				// 	continue
				// }

				// application -----------------------------------------------
				name := entry.Process.Name
				args, err := utils.GetCmd(entry.Process.Pid)
				if err == nil && len(args) > 0 {
					name = args[0]
					args = args[1:]
				}

				app := &models.Application{Name: name, Args: args, MachineID: machine.ID}
				pid := entry.Process.Pid
				if pid >= 0 {
					app.PID = uint64(pid)
				}

				hApp := hashApplication(app)
				if existingApp, exists := uniqueApps[hApp]; exists {
					app = existingApp
				} else {
					uniqueApps[hApp] = app
				}
				// apps = append(apps, app)

				// detect flow direction (we know that we are not in the listen case here)
				if isListening(entry.LocalAddr.String()) {
					// incoming flow

					// create all the endpoints (normally only one)
					for _, nic := range ipMapper[entry.LocalAddr.IP.String()] {
						flow := models.Flow{}
						if nic == nil {
							continue
						}
						endpoint := &models.ApplicationEndpoint{
							Addr:                 entry.LocalAddr.IP.String(),
							Port:                 entry.LocalAddr.Port,
							Protocol:             netstatProtocols[k],
							Application:          app,
							NetworkInterfaceID:   nic.ID,
							ApplicationProtocols: []string{},
						}
						// endpoints = append(endpoints, &endpoint)
						h := hashEndpoint(endpoint)
						if e, exists := uniqueEnpoints[h]; exists {
							endpoint = e
						} else {
							uniqueEnpoints[h] = endpoint
						}

						// for the moment we do not create the remote nic
						// the remote addr is the source of the flow
						flow.SrcAddr = entry.RemoteAddr.IP.String()
						flow.DstEndpoint = endpoint
						flows = append(flows, &flow)
					}

				} else {
					// outgoing flow
					for _, nic := range ipMapper[entry.LocalAddr.IP.String()] {
						if nic == nil {
							continue
						}
						flow := models.Flow{}

						// for the moment we do not create the remote nic
						// the remote addr is the destination of the flow
						endpoint := &models.ApplicationEndpoint{
							Addr:     entry.RemoteAddr.IP.String(),
							Port:     entry.RemoteAddr.Port,
							Protocol: netstatProtocols[k],
							// Application: ???, we do not know which app is running there
						}
						// we need to detect the remote nic if it exists to avoid duplicates
						remoteNIC, exists := ipNICMap[entry.RemoteAddr.IP.String()]
						if exists && remoteNIC != nil {
							endpoint.NetworkInterfaceID = remoteNIC.ID
						}

						h := hashEndpoint(endpoint)
						if e, exists := uniqueEnpoints[h]; exists {
							endpoint = e
						} else {
							uniqueEnpoints[h] = endpoint
						}

						flow.DstEndpoint = endpoint
						flow.SrcAddr = entry.LocalAddr.IP.String()
						flow.SrcNetworkInterface = nic
						flow.SrcNetworkInterfaceID = nic.ID
						flow.SrcApplication = app

						flows = append(flows, &flow)
					}
				}

			}
		}
	}

	// put applications in a slice
	for _, app := range uniqueApps {
		apps = append(apps, app)
	}
	// create or update applications
	if len(apps) > 0 {
		err = storage.DB().NewInsert().Model(&apps).
			On("CONFLICT (machine_id, name, pid) DO UPDATE").
			Set("updated_at = CURRENT_TIMESTAMP").
			Scan(ctx)
		if err != nil {
			return fmt.Errorf("fail to insert applications: %w", err)
		}
	}

	// put user-apps in a slice
	for _, ua := range uniqueUserApps {
		if ua.Application != nil {
			ua.ApplicationID = ua.Application.ID
		}
		if ua.User != nil {
			ua.UserID = ua.User.ID
		}
		userApps = append(userApps, ua)
	}
	// create or update user-apps
	if len(userApps) > 0 {
		err = storage.DB().NewInsert().Model(&userApps).
			On("CONFLICT (user_id, application_id) DO UPDATE").
			Set("updated_at = CURRENT_TIMESTAMP").
			Scan(ctx)
		if err != nil {
			return fmt.Errorf("fail to insert user-applications: %w", err)
		}
	}

	// put endpoints in a slice
	for _, endpoint := range uniqueEnpoints {
		if endpoint.Application != nil {
			endpoint.ApplicationID = endpoint.Application.ID
		}
		endpoints = append(endpoints, endpoint)
	}
	// create or update endpoints
	if len(endpoints) > 0 {
		err = storage.DB().NewInsert().Model(&endpoints).
			On("CONFLICT (port, protocol, addr, network_interface_id) DO UPDATE").
			Set("updated_at = CURRENT_TIMESTAMP").
			Scan(ctx)
		if err != nil {
			return fmt.Errorf("fail to insert application endpoints: %w", err)
		}
	}

	// set endpoint IDs in flows and deduplicate
	for _, flow := range flows {
		if flow.DstEndpoint != nil {
			flow.DstEndpointID = flow.DstEndpoint.ID
		}
		if flow.SrcApplication != nil {
			flow.SrcApplicationID = flow.SrcApplication.ID
		}
		// deduplicate flows based on unique constraint
		h := hashFlow(flow)
		if _, exists := uniqueFlows[h]; !exists {
			uniqueFlows[h] = flow
		}
	}

	// collect unique flows into slice
	flows = make([]*models.Flow, 0, len(uniqueFlows))
	for _, flow := range uniqueFlows {
		flows = append(flows, flow)
	}
	// create flows
	if len(flows) > 0 {
		err = storage.DB().NewInsert().Model(&flows).
			On("CONFLICT (src_application_id, src_addr, dst_endpoint_id) DO UPDATE").
			Set("updated_at = CURRENT_TIMESTAMP").
			Scan(ctx)
		if err != nil {
			return fmt.Errorf("fail to insert flows: %w", err)
		}
	}
	logger.
		WithField("flows", len(flows)).
		WithField("applications", len(apps)).
		WithField("endpoints", len(endpoints)).
		Info("Netstat scan completed")

	return nil
}
