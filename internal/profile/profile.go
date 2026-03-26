package profile

import "fmt"

// Profile represents a set of HTTP headers that simulate a particular client identity.
type Profile struct {
	Name    string
	Headers map[string]string
}

// GetProfile returns a built-in profile by name.
func GetProfile(name string) (*Profile, error) {
	p, ok := builtinProfiles[name]
	if !ok {
		return nil, fmt.Errorf("unknown profile %q — available profiles: browser-like, bot-like, crawler-like", name)
	}
	return &p, nil
}

// AvailableProfiles returns the names of all built-in profiles.
func AvailableProfiles() []string {
	names := make([]string, 0, len(builtinProfiles))
	for name := range builtinProfiles {
		names = append(names, name)
	}
	return names
}
