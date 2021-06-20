package rn2483

import (
	"fmt"
	"regexp"
	"time"
)

// Sleep causes the device to sleep for the specified duration
func (d *Device) Sleep(duration time.Duration) error {
	return d.ExecuteCommandCheckedStrict("sys sleep %d", duration.Milliseconds())
}

// GetHWEUI gets the device's preprogrammed EUI node address as a hex string
func (d *Device) GetHWEUI() (string, error) {
	return d.ExecuteCommandChecked("sys get hweui")
}

var (
	// FirmwareVersionRegex is a regex for breaking down a version string reported by the device
	FirmwareVersionRegex = regexp.MustCompile("^(.+)\\s+([0-9]+\\.[0-9]+\\.[0-9]+)\\s+(.+)$")
)

// DeviceSKU identifies a device model
type DeviceSKU string
const (
	DeviceRN2483 DeviceSKU = "RN2483"
	DeviceRN2903 DeviceSKU = "RN2903"
)
// KnownDeviceSKUs is a list of all known device SKUs
var KnownDeviceSKUs = []DeviceSKU{DeviceRN2483, DeviceRN2903}

// FirmwareVersion contains the result of parsing a version string returned by several sys commands
type FirmwareVersion struct {
	// Raw contains the original version string
	Raw string

	// SKU indicates the device model
	SKU DeviceSKU

	// Major is the first component of the firmware version number
	Major    int
	// Minor is the second component of the firmware version number
	Minor    int
	// Revision is the third component of the firmware version number
	Revision int

	// ReleaseTime is the time at which this firmware version was released
	ReleaseTime time.Time
}

// VersionString combines the major, minor and revision components into a single string for easier printing
func (fw *FirmwareVersion) VersionString() string {
	return fmt.Sprintf("%d.%d.%d", fw.Major, fw.Minor, fw.Revision)
}

// IsKnownSKU compares this firmware version's reported SKU to the list of known SKUs
func (fw *FirmwareVersion) IsKnownSKU() bool {
	for _, sku := range KnownDeviceSKUs {
		if fw.SKU == sku {
			return true
		}
	}
	return false
}

// ParseFirmwareVersion attempts to parse a version string reported by the device
func ParseFirmwareVersion(s string) (*FirmwareVersion, error) {
	matches := FirmwareVersionRegex.FindStringSubmatch(s)

	const expectedMatchCount = 4
	if len(matches) != expectedMatchCount {
		return nil, fmt.Errorf("firmware version regex did not match, expected %d groups, got %d", expectedMatchCount, len(matches))
	}

	v := &FirmwareVersion{
		Raw: s,
		SKU: DeviceSKU(matches[1]),
	}
	versionString := matches[2]
	dateString := matches[3]

	_, err := fmt.Sscanf(versionString, "%d.%d.%d", &v.Major, &v.Minor, &v.Revision)
	if err != nil {
		return nil, fmt.Errorf("error parsing version numbers: %w", err)
	}

	v.ReleaseTime, err = time.Parse("Jan 02 2006 15:04:05", dateString)
	if err != nil {
		return nil, fmt.Errorf("error parsing release date: %w", err)
	}

	return v, nil
}

// Reset resets the device, reverting to its stored configuration
func (d *Device) Reset() (*FirmwareVersion, error) {
	return d.executeCommandReturningFirmwareVersion("sys reset")
}

// FactoryReset resets the device to its factory configuration
func (d *Device) FactoryReset() (*FirmwareVersion, error) {
	return d.executeCommandReturningFirmwareVersion("sys factoryRESET")
}

// GetVersion retrieves the device's version information
func (d *Device) GetVersion() (*FirmwareVersion, error) {
	return d.executeCommandReturningFirmwareVersion("sys get ver")
}

func (d *Device) executeCommandReturningFirmwareVersion(format string, a ...interface{}) (*FirmwareVersion, error) {
	line, err := d.ExecuteCommandChecked(format, a...)
	if err != nil {
		return nil, err
	}

	return ParseFirmwareVersion(line)
}
