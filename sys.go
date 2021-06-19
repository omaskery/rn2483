package rn2483

import (
	"fmt"
	"regexp"
	"time"
)

func (d *Device) Sleep(duration time.Duration) error {
	return d.ExecuteCommandChecked("sys sleep %d", duration.Milliseconds())
}

func (d *Device) GetHWEUI() (string, error) {
	return d.ExecuteCommand("sys get hweui")
}

var (
	FirmwareVersionRegex = regexp.MustCompile("^(.+)\\s+([0-9]+\\.[0-9]+\\.[0-9]+)\\s+(.+)$")
)

type DeviceSKU string
const (
	DeviceRN2483 DeviceSKU = "RN2483"
	DeviceRN2903 DeviceSKU = "RN2903"
)
var KnownDeviceSKUs = []DeviceSKU{DeviceRN2483, DeviceRN2903}

type FirmwareVersion struct {
	Raw string

	SKU DeviceSKU

	Major    int
	Minor    int
	Revision int

	ReleaseTime time.Time
}

func (fw *FirmwareVersion) VersionString() string {
	return fmt.Sprintf("%d.%d.%d", fw.Major, fw.Minor, fw.Revision)
}

func (fw *FirmwareVersion) IsKnownSKU() bool {
	for _, sku := range KnownDeviceSKUs {
		if fw.SKU == sku {
			return true
		}
	}
	return false
}

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

func (d *Device) Reset() (*FirmwareVersion, error) {
	return d.executeCommandReturningFirmwareVersion("sys reset")
}

func (d *Device) FactoryReset() (*FirmwareVersion, error) {
	return d.executeCommandReturningFirmwareVersion("sys factoryRESET")
}

func (d *Device) GetVersion() (*FirmwareVersion, error) {
	return d.executeCommandReturningFirmwareVersion("sys get ver")
}

func (d *Device) executeCommandReturningFirmwareVersion(format string, a ...interface{}) (*FirmwareVersion, error) {
	line, err := d.ExecuteCommand(format, a...)
	if err != nil {
		return nil, err
	}

	return ParseFirmwareVersion(line)
}
