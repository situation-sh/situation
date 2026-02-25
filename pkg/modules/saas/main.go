package saas

import (
	"errors"
	"fmt"

	"github.com/situation-sh/situation/pkg/models"
)

var detectors = map[string]SaaSDetector{}

func registerDetector(detector SaaSDetector) {
	if _, exists := detectors[detector.Name()]; !exists {
		detectors[detector.Name()] = detector
	} else {
		msg := fmt.Sprintf("Detector with name %s is already registered", detector.Name())
		panic(msg)
	}
}

func Detect(endpoint *models.ApplicationEndpoint) (bool, string, error) {
	errs := []error{}
	for name, detector := range detectors {
		match, err := detector.Detect(endpoint)
		if err != nil {
			errs = append(errs, fmt.Errorf("SaaS detector %s: %w", name, err))
			continue
		}
		if match {
			// First match wins
			endpoint.SaaS = name
			return true, name, nil
		}
	}

	return false, "", errors.Join(errs...)
}
