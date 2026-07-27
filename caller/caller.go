// Package caller defines provider-neutral request caller labels.
//
// A caller is not an authentication scheme. It is a small application-defined
// label such as "public-api", "dashboard", "worker", or "admin" that other
// httpapi packages can use to describe which request sources may reach an
// endpoint or supply a parameter.
package caller

import (
	"encoding/json"
	"fmt"
	"strings"
	"unicode"
)

// Caller identifies an application-defined request source.
//
// Caller values are immutable and comparable. Define callers at package level
// and pass those values to endpoint groups, endpoints, request contexts, and
// parameter availability rules.
type Caller struct {
	name string
}

// Define returns a caller with name.
//
// Empty names panic because caller availability is part of the request contract
// and should fail at startup when misconfigured.
func Define(name string) Caller {
	caller, err := Parse(name)
	if err != nil {
		panic(err.Error())
	}
	return caller
}

// Parse returns a caller with name, or an error when name is not a valid caller
// label.
func Parse(name string) (Caller, error) {
	name = normalizeName(name)
	if name == "" {
		return Caller{}, fmt.Errorf("httpapi/caller: caller name is required")
	}
	for _, r := range name {
		if unicode.IsSpace(r) || unicode.IsControl(r) {
			return Caller{}, fmt.Errorf("httpapi/caller: caller name %q contains whitespace or control characters", name)
		}
	}
	return Caller{name: name}, nil
}

// Name returns the caller's stable name.
func (caller Caller) Name() string {
	return caller.name
}

// Defined reports whether caller was created with Define.
func (caller Caller) Defined() bool {
	return caller.name != ""
}

// IsZero reports whether caller is undefined.
func (caller Caller) IsZero() bool {
	return !caller.Defined()
}

// String returns the caller name.
func (caller Caller) String() string {
	return caller.name
}

// MarshalText encodes caller as its name.
func (caller Caller) MarshalText() ([]byte, error) {
	if !caller.Defined() {
		return nil, fmt.Errorf("httpapi/caller: cannot marshal undefined caller")
	}
	return []byte(caller.name), nil
}

// MarshalJSON encodes caller as its name.
func (caller Caller) MarshalJSON() ([]byte, error) {
	if !caller.Defined() {
		return []byte("null"), nil
	}
	return json.Marshal(caller.name)
}

// Set is a normalized caller allow-list.
//
// The zero value is unrestricted: every caller is allowed.
type Set struct {
	restricted bool
	callers    []Caller
	index      map[string]Caller
}

// SetOf returns a normalized caller set.
func SetOf(callers ...Caller) Set {
	if len(callers) == 0 {
		return Set{}
	}

	set := Set{
		restricted: true,
		callers:    make([]Caller, 0, len(callers)),
		index:      make(map[string]Caller, len(callers)),
	}
	for _, caller := range callers {
		caller = requireDefined(caller)
		if _, ok := set.index[caller.name]; ok {
			panic(fmt.Sprintf("httpapi/caller: duplicate caller %q", caller.name))
		}
		set.index[caller.name] = caller
		set.callers = append(set.callers, caller)
	}
	return set
}

// AvailableTo returns a restricted caller set containing callers.
func AvailableTo(callers ...Caller) Set {
	if len(callers) == 0 {
		panic("httpapi/caller: at least one caller is required")
	}
	return restrictedSet(callers...)
}

// Restricted reports whether this set restricts callers.
func (set Set) Restricted() bool {
	return set.restricted
}

// Allows reports whether caller is included in this set.
//
// An unrestricted set allows every caller, including the zero caller.
func (set Set) Allows(caller Caller) bool {
	if !set.Restricted() {
		return true
	}
	if !caller.Defined() {
		return false
	}
	_, ok := set.index[caller.name]
	return ok
}

// Callers returns a copy of the callers in definition order.
func (set Set) Callers() []Caller {
	if len(set.callers) == 0 {
		return nil
	}
	out := make([]Caller, len(set.callers))
	copy(out, set.callers)
	return out
}

// Intersect returns the callers allowed by both sets.
//
// The zero set is unrestricted, so intersecting with it returns the other set.
func (set Set) Intersect(other Set) Set {
	if !set.Restricted() {
		return other
	}
	if !other.Restricted() {
		return set
	}

	callers := make([]Caller, 0, len(set.callers))
	for _, candidate := range set.callers {
		if other.Allows(candidate) {
			callers = append(callers, candidate)
		}
	}
	return restrictedSet(callers...)
}

func requireDefined(caller Caller) Caller {
	if !caller.Defined() {
		panic("httpapi/caller: undefined caller")
	}
	return caller
}

func restrictedSet(callers ...Caller) Set {
	if len(callers) == 0 {
		return Set{restricted: true, index: map[string]Caller{}}
	}
	set := SetOf(callers...)
	set.restricted = true
	return set
}

func normalizeName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}
