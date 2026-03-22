package config

import (
	"fmt"
	"sort"
)

// ListProfiles returns sorted profile names.
func (c *Config) ListProfiles() []string {
	names := make([]string, 0, len(c.Profiles))
	for name := range c.Profiles {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// AddProfile adds or replaces a profile.
func (c *Config) AddProfile(key string, profile Profile) {
	if c.Profiles == nil {
		c.Profiles = make(map[string]Profile)
	}
	c.Profiles[key] = profile
}

// RemoveProfile deletes a profile. Returns error if it's the active one.
func (c *Config) RemoveProfile(key string) error {
	if key == c.ActiveProfile {
		return fmt.Errorf("cannot remove the active profile %q — switch to another profile first", key)
	}
	delete(c.Profiles, key)
	return nil
}

// SetActiveProfile switches the active profile.
func (c *Config) SetActiveProfile(key string) error {
	if _, ok := c.Profiles[key]; !ok {
		return fmt.Errorf("profile %q does not exist", key)
	}
	c.ActiveProfile = key
	return nil
}
