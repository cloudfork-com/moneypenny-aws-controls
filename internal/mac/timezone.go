package mac

import "time"

var userLocation *time.Location = time.Local

// SetTimezone sets the user's timezone.
func SetTimezone(tz string) error {
	if tz == "" {
		return nil
	}
	loc, err := time.LoadLocation(tz)
	if err == nil {
		userLocation = loc
	}
	return err
}
