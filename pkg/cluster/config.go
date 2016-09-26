// Copyright 2015 Sorint.lab
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied
// See the License for the specific language governing permissions and
// limitations under the License.

package cluster

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const (
	DefaultProxyCheckInterval = 5 * time.Second

	DefaultRequestTimeout          = 10 * time.Second
	DefaultSleepInterval           = 5 * time.Second
	DefaultKeeperFailInterval      = 20 * time.Second
	DefaultMaxStandbysPerSender    = 3
	DefaultSynchronousReplication  = false
	DefaultInitWithMultipleKeepers = false
	DefaultUsePGRewind             = false
)

// NilConfig is the cluster configuration with all the values as pointer. Having
// all the value as pointers is needed to know if, after json unmarshalling, the
// field wasn't provided. Not using a pointer will be impossible to know if a
// field having the go default value was provided or not.
// This is needed for different things:
// * Save in the cluster view only the user changed values. So if a default
// value is changed in future stolon versions it'll be automatically reflected.
// * Patching config
type NilConfig struct {
	RequestTimeout          *Duration          `json:"request_timeout,omitempty"`
	SleepInterval           *Duration          `json:"sleep_interval,omitempty"`
	KeeperFailInterval      *Duration          `json:"keeper_fail_interval,omitempty"`
	MaxStandbysPerSender    *uint              `json:"max_standbys_per_sender,omitempty"`
	SynchronousReplication  *bool              `json:"synchronous_replication,omitempty"`
	InitWithMultipleKeepers *bool              `json:"init_with_multiple_keepers,omitempty"`
	UsePGRewind             *bool              `json:"use_pg_rewind,omitempty"`
	PGParameters            *map[string]string `json:"pg_parameters,omitempty"`
}

// Config is the cluster configuration taken from a NilConfig and populated with
// all the defaults
type Config struct {
	// Time after which any request (keepers checks from sentinel etc...) will fail.
	RequestTimeout time.Duration
	// Interval to wait before next check (for every component: keeper, sentinel, proxy).
	SleepInterval time.Duration
	// Interval after the first fail to declare a keeper as not healthy.
	KeeperFailInterval time.Duration
	// Max number of standbys for every sender. A sender can be a master or
	// another standby (with cascading replication).
	MaxStandbysPerSender uint
	// Use Synchronous replication between master and its standbys
	SynchronousReplication bool
	// Choose a random initial master when multiple keeper are registered
	InitWithMultipleKeepers bool
	// Whether to use pg_rewind
	UsePGRewind bool
	// Map of postgres parameters
	PGParameters map[string]string
}

// StringP is a helper function that returns the address of a copy of the
// provided string (since the argument is passed by value).
func StringP(s string) *string {
	return &s
}

// UintP is a helper function that returns the address of a copy of the
// provided uint (since the argument is passed by value).
func UintP(u uint) *uint {
	return &u
}

// BoolP is a helper function that returns the address of a copy of the
// provided bool (since the argument is passed by value).
func BoolP(b bool) *bool {
	return &b
}

// DurationP is a helper function that returns the address of a copy of the
// provided Duration (since the argument is passed by value).
func DurationP(d Duration) *Duration {
	return &d
}

// MapStringP is a helper function that returns the address of a copy of the
// provided map[string]string.
func MapStringP(m map[string]string) *map[string]string {
	nm := map[string]string{}
	for k, v := range m {
		nm[k] = v
	}
	return &nm
}

// UnmarshalJSON implements the encoding/json.Unmarshaler interface. After
// unmarshalling it also validates the NilConfig.
func (c *NilConfig) UnmarshalJSON(in []byte) error {
	// nilConfig is needed to avoid recursive infinite calls to
	// NilConfig.UnmarshalJSON
	type nilConfig NilConfig
	var nc nilConfig
	if err := json.Unmarshal(in, &nc); err != nil {
		return err
	}
	*c = NilConfig(nc)
	if err := c.Validate(); err != nil {
		return fmt.Errorf("config validation failed: %v", err)
	}
	return nil
}

// Copy returns a deep copy
func (c *NilConfig) Copy() *NilConfig {
	if c == nil {
		return c
	}
	var nc NilConfig
	if c.RequestTimeout != nil {
		nc.RequestTimeout = DurationP(*c.RequestTimeout)
	}
	if c.SleepInterval != nil {
		nc.SleepInterval = DurationP(*c.SleepInterval)
	}
	if c.KeeperFailInterval != nil {
		nc.KeeperFailInterval = DurationP(*c.KeeperFailInterval)
	}
	if c.MaxStandbysPerSender != nil {
		nc.MaxStandbysPerSender = UintP(*c.MaxStandbysPerSender)
	}
	if c.SynchronousReplication != nil {
		nc.SynchronousReplication = BoolP(*c.SynchronousReplication)
	}
	if c.InitWithMultipleKeepers != nil {
		nc.InitWithMultipleKeepers = BoolP(*c.InitWithMultipleKeepers)
	}
	if c.UsePGRewind != nil {
		nc.UsePGRewind = BoolP(*c.UsePGRewind)
	}
	if c.PGParameters != nil {
		nc.PGParameters = MapStringP(*c.PGParameters)
	}
	return &nc
}

// Copy returns a deep copy
func (c *Config) Copy() *Config {
	if c == nil {
		return c
	}
	nc := *c
	return &nc
}

// Duration is needed to be able to marshal/unmarshal json strings with time
// unit (eg. 3s, 100ms) instead of ugly times in nanoseconds.
type Duration struct {
	time.Duration
}

// MarshalJSON marshals the Duration in time units (eg. 3s, 100ms).
// Implements the encoding/json.Marshaler interface.
func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

// UnmarshalJSON unmarshals the Duration in time units (eg. 3s,
// 100ms). Implements the encoding/json.Unmarshaler interface.
func (d *Duration) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	du, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	d.Duration = du
	return nil
}

// Validate validates a NilConfig
func (c *NilConfig) Validate() error {
	if c.RequestTimeout != nil && (*c.RequestTimeout).Duration < 0 {
		return fmt.Errorf("request_timeout must be positive")
	}
	if c.SleepInterval != nil && (*c.SleepInterval).Duration < 0 {
		return fmt.Errorf("sleep_interval must be positive")
	}
	if c.KeeperFailInterval != nil && (*c.KeeperFailInterval).Duration < 0 {
		return fmt.Errorf("keeper_fail_interval must be positive")
	}
	if c.MaxStandbysPerSender != nil && *c.MaxStandbysPerSender < 1 {
		return fmt.Errorf("max_standbys_per_sender must be at least 1")
	}
	return nil
}

// MergeDefaults merges the default values
func (c *NilConfig) MergeDefaults() {
	if c.RequestTimeout == nil {
		c.RequestTimeout = &Duration{DefaultRequestTimeout}
	}
	if c.SleepInterval == nil {
		c.SleepInterval = &Duration{DefaultSleepInterval}
	}
	if c.KeeperFailInterval == nil {
		c.KeeperFailInterval = &Duration{DefaultKeeperFailInterval}
	}
	if c.MaxStandbysPerSender == nil {
		c.MaxStandbysPerSender = UintP(DefaultMaxStandbysPerSender)
	}
	if c.SynchronousReplication == nil {
		c.SynchronousReplication = BoolP(DefaultSynchronousReplication)
	}
	if c.InitWithMultipleKeepers == nil {
		c.InitWithMultipleKeepers = BoolP(DefaultInitWithMultipleKeepers)
	}
	if c.UsePGRewind == nil {
		c.UsePGRewind = BoolP(DefaultUsePGRewind)
	}
	if c.PGParameters == nil {
		c.PGParameters = &map[string]string{}
	}
}

// ToConfig returns a *Config from a *NilConfig (it'll be populated with all the
// default values
func (c *NilConfig) ToConfig() *Config {
	nc := c.Copy()
	nc.MergeDefaults()
	return &Config{
		RequestTimeout:          (*nc.RequestTimeout).Duration,
		SleepInterval:           (*nc.SleepInterval).Duration,
		KeeperFailInterval:      (*nc.KeeperFailInterval).Duration,
		MaxStandbysPerSender:    *nc.MaxStandbysPerSender,
		SynchronousReplication:  *nc.SynchronousReplication,
		InitWithMultipleKeepers: *nc.InitWithMultipleKeepers,
		UsePGRewind:             *nc.UsePGRewind,
		PGParameters:            *nc.PGParameters,
	}
}

// NewDefaultConfig returns a *Config populated with all the default values.
func NewDefaultConfig() *Config {
	nc := &NilConfig{}
	nc.MergeDefaults()
	return nc.ToConfig()
}
