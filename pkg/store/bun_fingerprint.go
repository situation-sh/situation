package store

import (
	"context"
	"fmt"
	"strings"

	"github.com/situation-sh/situation/pkg/models"
	"github.com/uptrace/bun/dialect"
)

// FingerprintQuery holds the local identifiers used to find a matching machine
type FingerprintQuery struct {
	Agent    string   // Agent UUID (definitive match)
	HostID   string   // System UUID (definitive match)
	MACs     []string // MAC addresses (fuzzy match)
	IPs      []string // IP addresses (fuzzy match)
	Hostname string   // Hostname (fuzzy match)
	Ports    []uint16 // Open ports (fuzzy match)
}

// FingerprintMatch represents a machine with its matching details
type FingerprintMatch struct {
	Machine      *models.Machine
	Score        float64
	IsDefinitive bool     // True if matched on Agent or HostID
	MatchedOn    []string // What attributes matched
}

// Fuzzy matching weights
const (
	WeightMAC      = 0.4 // Per matching MAC
	WeightIP       = 0.2 // Per matching IP
	WeightHostname = 0.2 // If hostname matches
	WeightPorts    = 0.1 // Per matching port (capped)

	// Minimum score to consider a fuzzy match valid
	MinFuzzyScore = 0.3

	// Maximum contribution from ports (to avoid port-only matches)
	MaxPortScore = 0.3
)

// FuzzyMatchResult holds the result of a fuzzy match query
type FuzzyMatchResult struct {
	MachineID     int64   `bun:"machine_id"`
	Score         float64 `bun:"score"`
	MACMatches    int     `bun:"mac_matches"`
	IPMatches     int     `bun:"ip_matches"`
	HostnameMatch int     `bun:"hostname_match"`
	PortMatches   int     `bun:"port_matches"`
}

// FindMachineByFingerprint searches for a machine matching the given fingerprint.
//
// Matching strategy:
//  1. If Agent matches → definitive match, return immediately
//  2. If HostID matches → definitive match, return immediately
//  3. Otherwise, compute fuzzy score based on MAC/IP/hostname/ports
//
// Returns nil if no match found or fuzzy score is below threshold.
func (s *BunStorage) FindMachineByFingerprint(ctx context.Context, query *FingerprintQuery) (*FingerprintMatch, error) {
	// 1. Try definitive match on Agent
	if query.Agent != "" {
		machine, err := s.findMachineByAgent(ctx, query.Agent)
		if err != nil {
			return nil, err
		}
		if machine != nil {
			return &FingerprintMatch{
				Machine:      machine,
				Score:        1.0,
				IsDefinitive: true,
				MatchedOn:    []string{"agent:" + query.Agent},
			}, nil
		}
	}

	// 2. Try definitive match on HostID
	if query.HostID != "" {
		machine, err := s.findMachineByHostID(ctx, query.HostID)
		if err != nil {
			return nil, err
		}
		if machine != nil {
			return &FingerprintMatch{
				Machine:      machine,
				Score:        1.0,
				IsDefinitive: true,
				MatchedOn:    []string{"host_id:" + query.HostID},
			}, nil
		}
	}

	// 3. Fuzzy matching on MAC/IP/hostname/ports
	return s.findMachineByFuzzyMatch(ctx, query)
}

// findMachineByAgent finds a machine by its agent UUID
func (s *BunStorage) findMachineByAgent(ctx context.Context, agent string) (*models.Machine, error) {
	if agent == "" {
		return nil, nil
	}

	machine := new(models.Machine)
	err := s.db.NewSelect().
		Model(machine).
		Where("agent = ?", agent).
		Relation("NICS").
		Limit(1).
		Scan(ctx)

	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, nil
		}
		s.onError(err)
		return nil, err
	}

	if machine.ID == 0 {
		return nil, nil
	}

	return machine, nil
}

// findMachineByHostID finds a machine by its system UUID
func (s *BunStorage) findMachineByHostID(ctx context.Context, hostID string) (*models.Machine, error) {
	if hostID == "" {
		return nil, nil
	}

	machine := new(models.Machine)
	err := s.db.NewSelect().
		Model(machine).
		Where("host_id = ?", hostID).
		Relation("NICS").
		Limit(1).
		Scan(ctx)

	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, nil
		}
		s.onError(err)
		return nil, err
	}

	if machine.ID == 0 {
		return nil, nil
	}

	return machine, nil
}

// findMachineByFuzzyMatch computes similarity scores directly in the database
// using subqueries for MAC, IP, hostname, and port matches
func (s *BunStorage) findMachineByFuzzyMatch(ctx context.Context, query *FingerprintQuery) (*FingerprintMatch, error) {
	if len(query.MACs) == 0 && len(query.IPs) == 0 && query.Hostname == "" && len(query.Ports) == 0 {
		return nil, nil
	}

	// Build the fuzzy match query based on dialect
	result, err := s.computeFuzzyMatchInDB(ctx, query)
	if err != nil {
		s.onError(err)
		return nil, err
	}

	if result == nil || result.Score < MinFuzzyScore {
		return nil, nil
	}

	// Load the full machine
	machine := new(models.Machine)
	err = s.db.NewSelect().
		Model(machine).
		Where("id = ?", result.MachineID).
		Relation("NICS").
		Scan(ctx)

	if err != nil {
		s.onError(err)
		return nil, err
	}

	// Build matchedOn list from the result
	matchedOn := make([]string, 0)
	if result.MACMatches > 0 {
		matchedOn = append(matchedOn, fmt.Sprintf("mac:%d", result.MACMatches))
	}
	if result.IPMatches > 0 {
		matchedOn = append(matchedOn, fmt.Sprintf("ip:%d", result.IPMatches))
	}
	if result.HostnameMatch > 0 {
		matchedOn = append(matchedOn, "hostname:"+query.Hostname)
	}
	if result.PortMatches > 0 {
		matchedOn = append(matchedOn, fmt.Sprintf("ports:%d", result.PortMatches))
	}

	return &FingerprintMatch{
		Machine:      machine,
		Score:        result.Score,
		IsDefinitive: false,
		MatchedOn:    matchedOn,
	}, nil
}

// computeFuzzyMatchInDB executes the fuzzy matching query in the database
func (s *BunStorage) computeFuzzyMatchInDB(ctx context.Context, query *FingerprintQuery) (*FuzzyMatchResult, error) {
	switch s.db.Dialect().Name() {
	case dialect.SQLite:
		return s.fuzzyMatchSQLite(ctx, query)
	case dialect.PG:
		return s.fuzzyMatchPostgres(ctx, query)
	default:
		return nil, fmt.Errorf("unsupported dialect for fuzzy matching")
	}
}

// fuzzyMatchSQLite computes fuzzy match scores in SQLite
func (s *BunStorage) fuzzyMatchSQLite(ctx context.Context, query *FingerprintQuery) (*FuzzyMatchResult, error) {
	// Build MAC values list for IN clause
	macPlaceholders := make([]string, len(query.MACs))
	macArgs := make([]any, len(query.MACs))
	for i, mac := range query.MACs {
		macPlaceholders[i] = "?"
		macArgs[i] = strings.ToLower(mac)
	}

	// Build port values list for IN clause
	portPlaceholders := make([]string, len(query.Ports))
	portArgs := make([]any, len(query.Ports))
	for i, port := range query.Ports {
		portPlaceholders[i] = "?"
		portArgs[i] = port
	}

	// Build IP matching conditions using json_each
	ipConditions := make([]string, len(query.IPs))
	ipArgs := make([]any, len(query.IPs))
	for i, ip := range query.IPs {
		ipConditions[i] = "EXISTS (SELECT 1 FROM json_each(ni.ip) WHERE value = ?)"
		ipArgs[i] = ip
	}

	// Build the query
	macInClause := "0"
	if len(query.MACs) > 0 {
		macInClause = fmt.Sprintf("LOWER(ni.mac) IN (%s)", strings.Join(macPlaceholders, ","))
	}

	ipMatchExpr := "0"
	if len(query.IPs) > 0 {
		ipMatchExpr = fmt.Sprintf("(%s)", strings.Join(ipConditions, " OR "))
	}

	portInClause := "0"
	if len(query.Ports) > 0 {
		portInClause = fmt.Sprintf("ae.port IN (%s)", strings.Join(portPlaceholders, ","))
	}

	sql := fmt.Sprintf(`
		SELECT
			m.id as machine_id,
			COALESCE(mac_count, 0) as mac_matches,
			COALESCE(ip_count, 0) as ip_matches,
			CASE WHEN LOWER(m.hostname) = ? THEN 1 ELSE 0 END as hostname_match,
			COALESCE(port_count, 0) as port_matches,
			(COALESCE(mac_count, 0) * ? +
			 COALESCE(ip_count, 0) * ? +
			 CASE WHEN LOWER(m.hostname) = ? THEN ? ELSE 0 END +
			 MIN(COALESCE(port_count, 0) * ?, ?)) as score
		FROM machines m
		LEFT JOIN (
			SELECT machine_id, COUNT(*) as mac_count
			FROM network_interfaces ni
			WHERE %s
			GROUP BY machine_id
		) mac_sub ON mac_sub.machine_id = m.id
		LEFT JOIN (
			SELECT ni.machine_id, COUNT(*) as ip_count
			FROM network_interfaces ni
			WHERE %s
			GROUP BY ni.machine_id
		) ip_sub ON ip_sub.machine_id = m.id
		LEFT JOIN (
			SELECT ni.machine_id, COUNT(DISTINCT ae.port) as port_count
			FROM network_interfaces ni
			JOIN application_endpoints ae ON ae.network_interface_id = ni.id
			WHERE %s
			GROUP BY ni.machine_id
		) port_sub ON port_sub.machine_id = m.id
		WHERE (COALESCE(mac_count, 0) + COALESCE(ip_count, 0) + COALESCE(port_count, 0) +
			   CASE WHEN LOWER(m.hostname) = ? THEN 1 ELSE 0 END) > 0
		HAVING score >= ?
		ORDER BY score DESC
		LIMIT 1
	`, macInClause, ipMatchExpr, portInClause)

	// Add hostname args in the right places
	finalArgs := make([]any, 0)
	finalArgs = append(finalArgs, macArgs...)                      // for mac IN clause
	finalArgs = append(finalArgs, ipArgs...)                       // for ip conditions
	finalArgs = append(finalArgs, portArgs...)                     // for port IN clause
	finalArgs = append(finalArgs, strings.ToLower(query.Hostname)) // for hostname_match
	finalArgs = append(finalArgs, WeightMAC)
	finalArgs = append(finalArgs, WeightIP)
	finalArgs = append(finalArgs, strings.ToLower(query.Hostname)) // for score calculation
	finalArgs = append(finalArgs, WeightHostname)
	finalArgs = append(finalArgs, WeightPorts, MaxPortScore)
	finalArgs = append(finalArgs, strings.ToLower(query.Hostname)) // for WHERE clause
	finalArgs = append(finalArgs, MinFuzzyScore)

	var result FuzzyMatchResult
	err := s.db.NewRaw(sql, finalArgs...).Scan(ctx, &result)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, nil
		}
		return nil, err
	}

	if result.MachineID == 0 {
		return nil, nil
	}

	return &result, nil
}

// fuzzyMatchPostgres computes fuzzy match scores in PostgreSQL
func (s *BunStorage) fuzzyMatchPostgres(ctx context.Context, query *FingerprintQuery) (*FuzzyMatchResult, error) {
	// Build the query using PostgreSQL array operations
	args := make([]any, 0)

	macArray := "{}"
	if len(query.MACs) > 0 {
		lowerMACs := make([]string, len(query.MACs))
		for i, mac := range query.MACs {
			lowerMACs[i] = strings.ToLower(mac)
		}
		macArray = "{" + strings.Join(lowerMACs, ",") + "}"
	}
	args = append(args, macArray)

	ipArray := "{}"
	if len(query.IPs) > 0 {
		ipArray = "{" + strings.Join(query.IPs, ",") + "}"
	}
	args = append(args, ipArray)

	portArray := "{}"
	if len(query.Ports) > 0 {
		portStrs := make([]string, len(query.Ports))
		for i, p := range query.Ports {
			portStrs[i] = fmt.Sprintf("%d", p)
		}
		portArray = "{" + strings.Join(portStrs, ",") + "}"
	}
	args = append(args, portArray)

	args = append(args, strings.ToLower(query.Hostname))
	args = append(args, WeightMAC, WeightIP, WeightHostname, WeightPorts, MaxPortScore)
	args = append(args, strings.ToLower(query.Hostname))
	args = append(args, MinFuzzyScore)

	sql := `
		SELECT
			m.id as machine_id,
			COALESCE(mac_sub.mac_count, 0) as mac_matches,
			COALESCE(ip_sub.ip_count, 0) as ip_matches,
			CASE WHEN LOWER(m.hostname) = $4 THEN 1 ELSE 0 END as hostname_match,
			COALESCE(port_sub.port_count, 0) as port_matches,
			(COALESCE(mac_sub.mac_count, 0) * $5 +
			 COALESCE(ip_sub.ip_count, 0) * $6 +
			 CASE WHEN LOWER(m.hostname) = $11 THEN $7 ELSE 0 END +
			 LEAST(COALESCE(port_sub.port_count, 0) * $8, $9)) as score
		FROM machines m
		LEFT JOIN (
			SELECT machine_id, COUNT(*) as mac_count
			FROM network_interfaces
			WHERE LOWER(mac) = ANY($1::text[])
			GROUP BY machine_id
		) mac_sub ON mac_sub.machine_id = m.id
		LEFT JOIN (
			SELECT machine_id, COUNT(*) as ip_count
			FROM network_interfaces
			WHERE ip && $2::text[]
			GROUP BY machine_id
		) ip_sub ON ip_sub.machine_id = m.id
		LEFT JOIN (
			SELECT ni.machine_id, COUNT(DISTINCT ae.port) as port_count
			FROM network_interfaces ni
			JOIN application_endpoints ae ON ae.network_interface_id = ni.id
			WHERE ae.port = ANY($3::int[])
			GROUP BY ni.machine_id
		) port_sub ON port_sub.machine_id = m.id
		WHERE (COALESCE(mac_sub.mac_count, 0) + COALESCE(ip_sub.ip_count, 0) + COALESCE(port_sub.port_count, 0) +
			   CASE WHEN LOWER(m.hostname) = $11 THEN 1 ELSE 0 END) > 0
		HAVING (COALESCE(mac_sub.mac_count, 0) * $5 +
			    COALESCE(ip_sub.ip_count, 0) * $6 +
			    CASE WHEN LOWER(m.hostname) = $11 THEN $7 ELSE 0 END +
			    LEAST(COALESCE(port_sub.port_count, 0) * $8, $9)) >= $12
		ORDER BY score DESC
		LIMIT 1
	`

	var result FuzzyMatchResult
	err := s.db.NewRaw(sql, args...).Scan(ctx, &result)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, nil
		}
		return nil, err
	}

	if result.MachineID == 0 {
		return nil, nil
	}

	return &result, nil
}
