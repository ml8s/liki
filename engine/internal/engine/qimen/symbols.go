package qimen

import "fmt"

// ParsePalaceIndex converts a Chinese gong name to a GongIndex value.
func ParsePalaceIndex(s string) (GongIndex, error) {
	for i := 1; i <= 9; i++ {
		if gongNames[i] == s {
			return GongIndex(i), nil
		}
	}
	return 0, fmt.Errorf("unknown gong: %q", s)
}

// ParseStarIndex converts a Chinese star name to a StarIndex value.
func ParseStarIndex(s string) (StarIndex, error) {
	for i := 1; i <= 9; i++ {
		if starNames[i] == s {
			return StarIndex(i), nil
		}
	}
	return 0, fmt.Errorf("unknown star: %q", s)
}

// ParseDoorIndex converts a Chinese door name to a DoorIndex value.
func ParseDoorIndex(s string) (DoorIndex, error) {
	for i := 1; i <= 8; i++ {
		if doorNames[i] == s {
			return DoorIndex(i), nil
		}
	}
	return 0, fmt.Errorf("unknown door: %q", s)
}
