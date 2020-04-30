// Copyright (c) 2016 Home Box Office, Inc. as an unpublished work. Neither
// this material nor any portion hereof may be copied or distributed without
// the express written consent of Home Box Office, Inc.
//
// This material also contains proprietary and confidential information
// of Home Box Office, Inc. and its suppliers, and may not be used by or
// disclosed to any person, in whole or in part, without the prior written
// consent of Home Box Office, Inc.

package strutil

// Set represents a hash set of strings. Note that the receivers for this
// type are non-pointer receivers, since a map is a reference type
type Set map[string]struct{}

// NewSet creates and returns a new Set
func NewSet(capacity int) Set {
	return make(Set, capacity)
}

// NewSetFromSlice creates and returns a new string set from the existing
// slice `s`
func NewSetFromSlice(vals []string) Set {
	s := make(Set, len(vals))
	s.Add(vals...)
	return s
}

// Add adds the specified `vals` to the string set
func (s Set) Add(vals ...string) {
	for _, v := range vals {
		s[v] = struct{}{}
	}
}

// Remove removes the specified `val` from the string set
func (s Set) Remove(val string) {
	delete(s, val)
}

// Contains returns true if the string set contains the specified `val`, false
// otherwise
func (s Set) Contains(val string) bool {
	_, ok := s[val]
	return ok
}

// Difference returns the set difference of s1 with s2 (s1 \ s2)
func (s Set) Difference(s2 Set) Set {
	res := make(Set)
	for k := range s {
		if _, ok := s2[k]; !ok {
			res.Add(k)
		}
	}
	return res
}

// Slice converts this Set to a string slice
func (s Set) Slice() []string {
	a := make([]string, 0, len(s))
	for k := range s {
		a = append(a, k)
	}
	return a
}
