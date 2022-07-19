package modules

import "fmt"

type MustBeRunAsRootError struct {
	UID string
}

func (m *MustBeRunAsRootError) Error() string {
	return fmt.Sprintf("It must be run as root (current uid: %s)", m.UID)
}
