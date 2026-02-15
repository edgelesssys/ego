//go:build !ego_fips140

package core

func checkFIPS140() error {
	return nil
}
