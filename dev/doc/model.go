package main

import (
	"go/doc"
	"strings"
)

type Status uint8

const (
	OK      = Status(0)
	KO      = Status(1)
	UNKNOWN = Status(2)
)

type ModuleDoc struct {
	Description string
	Data        []string
	OS          map[string]Status
	Arch        map[string]Status
	Comments    string
}

func StatusFromString(s string) Status {
	s = strings.TrimSpace(strings.ToLower(s))
	switch s {
	case "yes":
	case "ok":
		return OK
	case "no":
	case "ko":
	case "error":
		return KO
	default:
		return UNKNOWN
	}
	return UNKNOWN
}

func NewFromPackageDoc(p *doc.Package) *ModuleDoc {
	m := ModuleDoc{
		Data: make([]string, 0),
		OS:   make(map[string]Status),
		Arch: make(map[string]Status),
	}
	if d, ok := p.Notes["DESCRIPTION"]; ok && len(d) > 0 {
		m.Description = d[0].Body
	}
	if d, ok := p.Notes["DATA"]; ok && len(d) > 0 {
		for _, s := range strings.Split(d[0].Body, ",") {
			m.Data = append(m.Data, strings.TrimSpace(s))
		}
	}
	if d, ok := p.Notes["OS"]; ok && len(d) > 0 {
		for _, n := range d {
			m.OS[n.UID] = StatusFromString(n.Body)
		}
	}
	if d, ok := p.Notes["ARCH"]; ok && len(d) > 0 {
		for _, n := range d {
			m.Arch[n.UID] = StatusFromString(n.Body)
		}
	}
	return &m
}
