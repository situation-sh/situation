//go:build windows
// +build windows

package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v3"
)

const taskName = "Situation"

var regInfoProps = map[string]interface{}{
	"Author":      "situation.sh",
	"Description": "Situation is a lightweight single binary that collects network, system and application data",
	"Version":     Version,
	"Date":        time.Now().Format("2006-01-02T15:04:05"),
	"Source":      "situation.sh",
}

var settingsProps = map[string]interface{}{
	"AllowDemandStart":           true,   // Allow manual run
	"AllowHardTerminate":         true,   // Allow to kill the task if it does not stop when asked
	"Enabled":                    true,   // Enable the task
	"Hidden":                     false,  // Show the task in the UI
	"ExecutionTimeLimit":         "PT5M", // Max run time (ISO 8601 format, here 5 minutes)
	"RestartCount":               3,      // Number of restart if the task fails
	"RestartInterval":            "PT1M", // Time between restarts (ISO 8601 format, minimum is 1 minute)
	"MultipleInstances":          0,      // 0: Ignore new instance, 1: Parallel, 2: Queue, 3: Stop existing
	"RunOnlyIfNetworkAvailable":  false,  // Do not wait for network to run the task
	"StopIfGoingOnBatteries":     false,  // Do not stop the task if on batteries
	"DisallowStartIfOnBatteries": false,  // Allow to start the task if on batteries

}

// Run background agent with full system rights
var principalProps = map[string]interface{}{
	"LogonType":   5,           // Runs using a built-in service account (like SYSTEM, LOCAL SERVICE, NETWORK SERVICE). No password needed.
	"RunLevel":    1,           // 0: Least privilege, 1: Highest privilege of the user
	"DisplayName": "Situation", // Friendly name
	"UserId":      "SYSTEM",    // User to run the task as
}

func getTriggerProps() map[string]interface{} {
	now := time.Now()
	sb := time.Date(now.Year(), now.Month(), now.Day(), startTime.Hour(), startTime.Minute(), startTime.Second(), 0, now.Location())
	var triggerProps = map[string]interface{}{
		"Enabled":       true,
		"StartBoundary": sb.Format("2006-01-02T15:04:05"),
		"DaysInterval":  daysPeriod,
	}
	return triggerProps
}

func getActionProps(args []string) map[string]interface{} {
	exePath, err := os.Executable()
	if err != nil {
		exePath = "C:\\Program Files\\situation\\situation.exe"
	}

	return map[string]interface{}{
		"Path":             exePath,
		"Arguments":        strings.Join(args, " "),
		"WorkingDirectory": "C:\\temp", // Just a writable directory
	}
}

func getRepetitionProps() map[string]interface{} {
	if timePeriod == 0 {
		return map[string]interface{}{}
	}
	return map[string]interface{}{
		"Interval":          fmt.Sprintf("PT%dS", uint64(timePeriod.Seconds())), // ISO 8601 format, here every N seconds
		"StopAtDurationEnd": false,                                              // Do not stop the task at the end of the duration
		"Duration":          "P1D",                                              // Duration of the repetition (here 1 day)
	}
}

func taskConnect() (*ole.IDispatch, error) {
	unknown, err := oleutil.CreateObject("Schedule.Service")
	if err != nil {
		return nil, fmt.Errorf("fail to create Schedule.Service object: %w", err)
	}
	// COM components implement the IDispatch interface to
	// enable access by Automation clients, such as Visual Basic.
	service, err := unknown.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		return nil, fmt.Errorf("fail to get IDispatch interface: %w", err)
	}
	if _, err := oleutil.CallMethod(service, "Connect"); err != nil {
		service.Release()
		return nil, fmt.Errorf("fail to connect to schedule service: %w", err)
	}
	return service, nil
}

func putProperties(disp *ole.IDispatch, props map[string]interface{}) error {
	for k, v := range props {
		if _, err := oleutil.PutProperty(disp, k, v); err != nil {
			return fmt.Errorf("fail to put property %s: %w", k, err)
		}
	}
	return nil
}

func createTask(service *ole.IDispatch) (*ole.IDispatch, error) {
	taskDefDisp, err := oleutil.CallMethod(service, "NewTask", 0)
	if err != nil {
		return nil, fmt.Errorf("fail to create new task: %w", err)
	}

	taskDef := taskDefDisp.ToIDispatch()
	if taskDef == nil {
		return nil, fmt.Errorf("fail to get task definition IDispatch")
	}
	return taskDef, nil
}

func createTrigger(triggers *ole.IDispatch) (*ole.IDispatch, error) {
	// https://learn.microsoft.com/fr-fr/windows/win32/api/taskschd/ne-taskschd-task_trigger_type2
	trigger, err := oleutil.CallMethod(triggers, "Create", 2) // Daily trigger
	if err != nil {
		return nil, fmt.Errorf("fail to create trigger: %w", err)
	}
	triggerDisp := trigger.ToIDispatch()
	if triggerDisp == nil {
		return nil, fmt.Errorf("fail to get trigger IDispatch")
	}
	return triggerDisp, nil
}

func createAction(actions *ole.IDispatch) (*ole.IDispatch, error) {
	// https://learn.microsoft.com/fr-fr/windows/win32/api/taskschd/ne-taskschd-task_action_type
	actionDisp, err := oleutil.CallMethod(actions, "Create", 0) // Exec action
	if err != nil {
		return nil, fmt.Errorf("fail to create action: %w", err)
	}
	action := actionDisp.ToIDispatch()
	if action == nil {
		return nil, fmt.Errorf("fail to get action IDispatch")
	}
	return action, nil
}

func getTaskProps(task *ole.IDispatch, name string) (*ole.IDispatch, error) {
	prop, err := oleutil.GetProperty(task, name)
	if err != nil {
		return nil, fmt.Errorf("fail to get %s: %w", name, err)
	}
	iprop := prop.ToIDispatch()
	if iprop == nil {
		return nil, fmt.Errorf("fail to get %s IDispatch", name)
	}
	return iprop, nil
}

func getRoot(service *ole.IDispatch) (*ole.IDispatch, error) {
	// get root folder (for incoming registration)
	root, err := oleutil.CallMethod(service, "GetFolder", `\`)
	if err != nil {
		return nil, fmt.Errorf("fail to get root folder: %w", err)
	}
	rootDisp := root.ToIDispatch()
	if rootDisp == nil {
		return nil, fmt.Errorf("fail to get root folder IDispatch")
	}
	return rootDisp, nil
}

func runTaskCmd(ctx context.Context, cmd *cli.Command) error {
	// https://learn.microsoft.com/fr-fr/windows/win32/api/objbase/ne-objbase-coinit
	if err := ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED); err != nil {
		return fmt.Errorf("fail to initialize COM library: %w", err)
	}
	defer ole.CoUninitialize()

	logrus.Debugf("Connecting to Windows Task Scheduler")
	service, err := taskConnect()
	if err != nil {
		return err
	}
	defer service.Release()

	// get root folder (for incoming registration)
	logrus.Debugf("Getting root folder")
	root, err := getRoot(service)
	if err != nil {
		return err
	}
	defer root.Release()

	if uninstall {
		logrus.Infof("Deleting scheduled task %s", taskName)
		// Unregister the task if it exists
		_, err := oleutil.CallMethod(root, "DeleteTask", taskName, 0)
		if err != nil {
			return fmt.Errorf("fail to delete task: %w", err)
		}
		return nil
	}

	logrus.Debugf("Creating scheduled task definition")
	taskDef, err := createTask(service)
	if err != nil {
		return err
	}
	defer taskDef.Release()

	// registration info
	logrus.Debugf("Setting registration info")
	regInfo, err := getTaskProps(taskDef, "RegistrationInfo")
	if err != nil {
		return err
	}
	defer regInfo.Release()
	if err := putProperties(regInfo, regInfoProps); err != nil {
		return err
	}

	// principal
	logrus.Debugf("Setting principal info")
	principal, err := getTaskProps(taskDef, "Principal")
	if err != nil {
		return err
	}
	defer principal.Release()
	if err := putProperties(principal, principalProps); err != nil {
		return err
	}

	// settings
	logrus.Debugf("Setting taks settings")
	settings, err := getTaskProps(taskDef, "Settings")
	if err != nil {
		return err
	}
	defer settings.Release()
	if err := putProperties(settings, settingsProps); err != nil {
		return err
	}

	// triggers
	logrus.Debugf("Setting task triggers")
	triggers, err := getTaskProps(taskDef, "Triggers")
	if err != nil {
		return err
	}
	defer triggers.Release()
	trigger, err := createTrigger(triggers)
	if err != nil {
		return err
	}
	defer trigger.Release()
	if err := putProperties(trigger, getTriggerProps()); err != nil {
		return err
	}

	// repetition
	logrus.Debugf("Setting task repetition")
	repetition, err := getTaskProps(trigger, "Repetition")
	if err != nil {
		return err
	}
	defer repetition.Release()
	if err := putProperties(repetition, getRepetitionProps()); err != nil {
		return err
	}

	// actions
	logrus.Debugf("Setting task actions")
	actions, err := getTaskProps(taskDef, "Actions")
	if err != nil {
		return err
	}
	defer actions.Release()
	action, err := createAction(actions)
	if err != nil {
		return err
	}
	defer action.Release()
	args := getRunArgs(cmd)
	if err := putProperties(action, getActionProps(args)); err != nil {
		return err
	}

	// REGISTER
	// https://learn.microsoft.com/fr-fr/windows/win32/taskschd/taskfolder-registertaskdefinition
	logrus.Infof("Registering scheduled task %s", taskName)
	_, err = oleutil.CallMethod(
		root,
		"RegisterTaskDefinition",
		taskName,
		taskDef,
		0x06,                        // TASK_CREATE_OR_UPDATE
		nil,                         // userId (given in principal props)
		nil,                         // password (given in principal props)
		principalProps["LogonType"], // logonType
		"",                          // sddl ?
	)
	return err
}
