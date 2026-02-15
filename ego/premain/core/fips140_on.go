//go:build ego_fips140

package core

import (
	"crypto/fips140"
	"errors"
)

func checkFIPS140() error {
	if !fips140.Enabled() {
		return errors.New("FIPS 140 not enabled")
	}
	return nil
}
