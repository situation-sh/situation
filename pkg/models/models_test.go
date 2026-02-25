package models

import (
	"testing"

	"github.com/invopop/jsonschema"
)

func TestNetworkInterfaceJSONSchema(t *testing.T) {
	s := jsonschema.Reflect(&NetworkInterface{})
	// m := s.Properties.Get("mac")
	t.Logf("PROPS: %+v\n", s.Definitions["NetworkInterface"].Properties.Value("mac"))
}
