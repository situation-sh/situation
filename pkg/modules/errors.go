package modules

import "fmt"

// import "fmt"

// type mustBeRunAsRootError struct {
// 	uid string
// }

// func (err *mustBeRunAsRootError) Error() string {
// 	return fmt.Sprintf("It must be run as root (current uid: %s)", err.uid)
// }

// notApplicableError is raised when a module is not applicable
// to a given target
type notApplicableError struct {
	msg string
}

func (err *notApplicableError) Error() string {
	return fmt.Sprintf("cannot run: %s", err.msg)
}
