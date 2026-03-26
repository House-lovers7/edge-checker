package safety

import "fmt"

// CheckProduction ensures that production environment requires explicit opt-in.
func CheckProduction(environment string, allowProductionFlag bool) error {
	if environment == "production" && !allowProductionFlag {
		return fmt.Errorf("environment is %q but --allow-production flag was not set — execution blocked for safety", environment)
	}
	return nil
}
