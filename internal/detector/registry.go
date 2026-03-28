package detector

import (
	"fmt"
	"strings"

	"golang.org/x/sys/windows/registry"
)

// RegistryLookup queries the Windows registry for a value at the given key path.
// Format: "HKCU\Software\AppName\ValueName" — last component is the value name.
func RegistryLookup(keyPath string) (string, error) {
	parts := strings.SplitN(keyPath, `\`, 2)
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid registry path: %s", keyPath)
	}

	var rootKey registry.Key
	switch strings.ToUpper(parts[0]) {
	case "HKCU", "HKEY_CURRENT_USER":
		rootKey = registry.CURRENT_USER
	case "HKLM", "HKEY_LOCAL_MACHINE":
		rootKey = registry.LOCAL_MACHINE
	default:
		return "", fmt.Errorf("unsupported registry root: %s", parts[0])
	}

	remaining := parts[1]
	lastSep := strings.LastIndex(remaining, `\`)
	if lastSep < 0 {
		return "", fmt.Errorf("invalid registry path (no value name): %s", keyPath)
	}
	subKeyPath := remaining[:lastSep]
	valueName := remaining[lastSep+1:]

	key, err := registry.OpenKey(rootKey, subKeyPath, registry.QUERY_VALUE)
	if err != nil {
		return "", fmt.Errorf("opening registry key %s: %w", subKeyPath, err)
	}
	defer key.Close()

	val, _, err := key.GetStringValue(valueName)
	if err != nil {
		return "", fmt.Errorf("reading registry value %s: %w", valueName, err)
	}

	return val, nil
}
