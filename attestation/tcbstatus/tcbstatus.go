// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package tcbstatus

// Status is the status of the enclave's TCB level.
type Status uint

//go:generate stringer -type=Status

// TCB level status of the enclave.
const (
	UpToDate Status = iota
	OutOfDate
	Revoked
	ConfigurationNeeded
	OutOfDateConfigurationNeeded
	SWHardeningNeeded
	ConfigurationAndSWHardeningNeeded
	Unknown
)

// Explain returns a description of the TCB status.
func Explain(status Status) string {
	// https://api.portal.trustedservices.intel.com/documentation
	switch status {
	case UpToDate:
		return "TCB level of the SGX platform is up-to-date."
	case OutOfDate:
		return "TCB level of SGX platform is outdated."
	case Revoked:
		return "TCB level of SGX platform is revoked. The platform is not trustworthy."
	case ConfigurationNeeded:
		return "TCB level of the SGX platform is up-to-date but additional configuration of SGX platform may be needed."
	case OutOfDateConfigurationNeeded:
		return "TCB level of SGX platform is outdated and additional configuration of SGX platform may be needed."
	case SWHardeningNeeded:
		return "TCB level of the SGX platform is up-to-date but due to certain issues affecting the platform, additional SW Hardening in the attesting SGX enclaves may be needed."
	case ConfigurationAndSWHardeningNeeded:
		return "TCB level of the SGX platform is up-to-date but additional configuration for the platform and SW Hardening in the attesting SGX enclaves may be needed."
	}
	return "unknown status"
}
