package app

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
)

// deviceInfo describes a target device's autorun asset set.
type deviceInfo struct {
	// icoName is the filename to write the icon as at the destination root.
	icoName string
	// icoData is the raw bytes of the .ico file to embed.
	icoData []byte
	// autorunLabel is the label value to write in autorun.inf.
	autorunLabel string
}

//go:embed assets/devices/trimui-brick/black/Black.ico
var deviceAssets embed.FS

// knownDevices is the registry of supported --device values.
var knownDevices = map[string]deviceInfo{
	"trimui-brick": {
		icoName:      "Black.ico",
		icoData:      mustReadEmbedded("assets/devices/trimui-brick/black/Black.ico"),
		autorunLabel: "NextUI Brick",
	},
}

func mustReadEmbedded(path string) []byte {
	data, err := deviceAssets.ReadFile(path)
	if err != nil {
		panic(fmt.Sprintf("embedded asset missing: %s: %v", path, err))
	}
	return data
}

// applyDeviceAssets writes the autorun .ico and autorun.inf to dstRoot for the
// named device. Returns an error if the device name is unknown.
func applyDeviceAssets(dstRoot, device string, logf func(string, ...any)) error {
	info, ok := knownDevices[device]
	if !ok {
		return fmt.Errorf("unknown --device %q (supported: %s)", device, supportedDeviceNames())
	}

	icoPath := filepath.Join(dstRoot, info.icoName)
	logf("device %s: writing %s", device, icoPath)
	if err := os.WriteFile(icoPath, info.icoData, 0o644); err != nil {
		return fmt.Errorf("device asset write %s: %w", icoPath, err)
	}

	autorunContent := fmt.Sprintf("[autorun]\r\nicon  = %s\r\nlabel = %s\r\n", info.icoName, info.autorunLabel)
	autorunPath := filepath.Join(dstRoot, "autorun.inf")
	logf("device %s: writing %s", device, autorunPath)
	if err := os.WriteFile(autorunPath, []byte(autorunContent), 0o644); err != nil {
		return fmt.Errorf("device asset write %s: %w", autorunPath, err)
	}

	return nil
}

// supportedDeviceNames returns the sorted list of known device names for error messages.
func supportedDeviceNames() string {
	names := make([]string, 0, len(knownDevices))
	for k := range knownDevices {
		names = append(names, k)
	}
	if len(names) == 1 {
		return names[0]
	}
	return fmt.Sprintf("%v", names)
}
