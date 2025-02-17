// Copyright 2017 The Kubernetes Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package dataselect

import (
	"fmt"
	"strings"
	"time"
)

// ----------------------- Standard Comparable Types ------------------------
// These types specify how given value should be compared
// They all implement ComparableValueInterface
// You can convert basic types to these types to support auto sorting etc.
// If you cant find your type compare here you will have to implement it yourself :)

// StdComparableInt must be equal ints
type StdComparableInt int

// Compare compares with  the given value
func (v StdComparableInt) Compare(otherV ComparableValue) int {
	other := otherV.(StdComparableInt)
	return intsCompare(int(v), int(other))
}

// Contains returns if value is contained
func (v StdComparableInt) Contains(otherV ComparableValue) bool {
	return v.Compare(otherV) == 0
}

type String interface {
	String() string
}

func ToString(v interface{}) string {
	if strV, ok := v.(String); ok {
		return strV.String()
	}
	return fmt.Sprintf("%v", v)
}

// StdEqualString strings that must be exactly the same
type StdEqualString string

// Compare compares with  the given value
func (v StdEqualString) Compare(otherV ComparableValue) int {
	return strings.Compare(v.String(), ToString(otherV))
}

// Contains returns if value is contained
func (v StdEqualString) Contains(otherV ComparableValue) bool {
	return v.Compare(otherV) == 0
}

func (v StdEqualString) String() string {
	return string(v)
}

// StdComparableString only equal decoded UTF-8 values will return true
type StdComparableString string

// Compare compares with  the given value
func (v StdComparableString) Compare(otherV ComparableValue) int {
	return strings.Compare(v.String(), ToString(otherV))
}

// Contains returns if value is contained
func (v StdComparableString) Contains(otherV ComparableValue) bool {
	return strings.EqualFold(v.String(), ToString(otherV))
}

func (v StdComparableString) String() string {
	return string(v)
}

// StdLowerComparableString constructor for StdComparableString returning a lower case string
func StdLowerComparableString(val string) StdComparableString {
	return StdComparableString(strings.ToLower(val))
}

// StdCaseInSensitiveComparableString case insensitive wrapper of StdComparableString
type StdCaseInSensitiveComparableString string

// Compare compares with  the given value
func (v StdCaseInSensitiveComparableString) Compare(otherV ComparableValue) int {
	return strings.Compare(strings.ToLower(v.String()), strings.ToLower(ToString(otherV)))
}

// Contains returns if value is contained
func (v StdCaseInSensitiveComparableString) Contains(otherV ComparableValue) bool {
	return strings.Contains(strings.ToLower(v.String()), strings.ToLower(ToString(otherV)))
}

func (v StdCaseInSensitiveComparableString) String() string {
	return string(v)
}

// StdComparableRFC3339Timestamp takes RFC3339 Timestamp strings and compares them as TIMES. In case of time parsing error compares values as strings.
type StdComparableRFC3339Timestamp string

// Compare compares with  the given value
func (v StdComparableRFC3339Timestamp) Compare(otherV ComparableValue) int {
	other := ToString(otherV)
	// try to compare as timestamp (earlier = smaller)
	selfTime, err1 := time.Parse(time.RFC3339, v.String())
	otherTime, err2 := time.Parse(time.RFC3339, other)

	if err1 != nil || err2 != nil {
		// in case of timestamp parsing failure just compare as strings
		return strings.Compare(string(v), other)
	}
	return ints64Compare(selfTime.Unix(), otherTime.Unix())
}

// Contains returns if value is contained
func (v StdComparableRFC3339Timestamp) Contains(otherV ComparableValue) bool {
	return v.Compare(otherV) == 0
}

func (v StdComparableRFC3339Timestamp) String() string {
	return string(v)
}

// StdComparableTime time.Time implementation for ComperableValue
type StdComparableTime time.Time

// Compare compares with  the given value
func (v StdComparableTime) Compare(otherV ComparableValue) int {
	other := otherV.(StdComparableTime)
	return ints64Compare(time.Time(v).Unix(), time.Time(other).Unix())
}

// Contains returns if value is contained
func (v StdComparableTime) Contains(otherV ComparableValue) bool {
	return v.Compare(otherV) == 0
}

// StdExactString exact comparisson of strings using == operator for ComparableValue
type StdExactString string

// Compare compares with  the given value
func (v StdExactString) Compare(otherV ComparableValue) int {
	return strings.Compare(v.String(), ToString(otherV))
}

// Contains returns if value is contained
func (v StdExactString) Contains(otherV ComparableValue) bool {
	return v.String() == ToString(otherV)
}

func (v StdExactString) String() string {
	return string(v)
}

// Int comparison functions. Similar to strings.Compare.
func intsCompare(a, b int) int {
	if a > b {
		return 1
	} else if a == b {
		return 0
	}
	return -1
}

func ints64Compare(a, b int64) int {
	if a > b {
		return 1
	} else if a == b {
		return 0
	}
	return -1
}

// StdComparableLabel label implementation of ComparableValue.
// supports multiple values split by comma ","
type StdComparableLabel string

// Compare compares with  the given value
func (v StdComparableLabel) Compare(otherV ComparableValue) int {
	return strings.Compare(v.String(), ToString(otherV))
}

// Contains returns if value is contained, if ComparableValue contains != , will treat as not equal label value
// eg. v="category:Build,source:customer,lang:go", otherV="lang!:java" will return true
func (v StdComparableLabel) Contains(otherV ComparableValue) bool {
	other := ToString(otherV)

	if strings.Contains(other, "!:") {
		// if containes !; , treat it as !=
		split := strings.Split(v.String(), ",")
		if len(split) == 0 {
			return true
		}
		for _, s := range split {
			replaced := strings.ReplaceAll(other, "!:", ":")
			if strings.EqualFold(s, replaced) {
				return false
			}
		}
		return true
	}

	split := strings.Split(v.String(), ",")
	if len(split) == 0 {
		return false
	}
	for _, s := range split {
		if strings.ToLower(s) == strings.ToLower(other) {
			return true
		}
	}
	return false
}

func (v StdComparableLabel) String() string {
	return string(v)
}

// StdComparableStringIn comparable string in.
// Supports multiple values using ":" as a separator
// if any of the values is equal returns true
type StdComparableStringIn string

// Compare compares with  the given value
func (v StdComparableStringIn) Compare(otherV ComparableValue) int {
	return strings.Compare(v.String(), ToString(otherV))
}

// Contains returns if value is contained
func (v StdComparableStringIn) Contains(otherV ComparableValue) bool {
	cur := string(v)
	split := strings.Split(ToString(otherV), ":")
	if len(split) == 0 {
		return true
	}
	for _, s := range split {
		if s == cur {
			return true
		}
	}
	return false
}

func (v StdComparableStringIn) String() string {
	return string(v)
}

// StdComparableContainsString if one string contains the other
type StdComparableContainsString string

// Compare compares with  the given value
func (v StdComparableContainsString) Compare(otherV ComparableValue) int {
	return strings.Compare(v.String(), ToString(otherV))
}

// Contains returns if value is contained
func (v StdComparableContainsString) Contains(otherV ComparableValue) bool {
	return strings.Contains(strings.ToLower(v.String()), strings.ToLower(ToString(otherV)))
}

func (v StdComparableContainsString) String() string {
	return string(v)
}

// MutilComparableValue provide more condition to compare and contion
type MutilComparableValue struct {
	Items []ComparableValue
}

// Compare compares with  the given value
// Compare will used in sort method
//
//	a := d.GenericDataList[i].GetProperty(sortBy.Property)
//	b := d.GenericDataList[j].GetProperty(sortBy.Property)
//	cmp := a.Compare(b)
//
// a and b 'type is  MutilComparableValue
func (mutil MutilComparableValue) Compare(otherV ComparableValue) int {
	others := otherV.(MutilComparableValue).Items
	for index, item := range mutil.Items {
		other := others[index]
		if item.Compare(other) != 0 {
			return item.Compare(other)
		}
	}
	return 0
}

// Contains find if one of the item contains otherV,if so return true
func (mutil MutilComparableValue) Contains(otherV ComparableValue) bool {
	for _, item := range mutil.Items {
		if item.Contains(otherV) {
			return true
		}
	}
	return false
}
