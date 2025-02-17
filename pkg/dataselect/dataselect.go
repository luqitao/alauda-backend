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
	"sort"
	"strings"
)

// DataCell describes the interface of the data cell that contains all the necessary methods needed to perform
// complex data selection
// GenericDataSelect takes a list of these interfaces and performs selection operation.
// Therefore as long as the list is composed of GenericDataCells you can perform any data selection!
type DataCell interface {
	// GetPropertyAtIndex returns the property of this data cell.
	// Value returned has to have Compare method which is required by Sort functionality of DataSelect.
	GetProperty(PropertyName) ComparableValue
}

// ComparableValue hold any value that can be compared to its own kind.
type ComparableValue interface {
	// Compares d with other value. Returns 1 if other value is smaller, 0 if they are the same, -1 if other is larger.
	Compare(ComparableValue) int
	// Returns true if d value contains or is equal to other value, false otherwise.
	Contains(ComparableValue) bool
}

// DataSelector contains all the required data to perform data selection.
// It implements sort.Interface so its sortable under sort.Sort
// You can use its Select method to get selected GenericDataCell list.
type DataSelector struct {
	// GenericDataList hold generic data cells that are being selected.
	GenericDataList []DataCell
	// Query holds instructions for data select.
	Query *Query
}

// Implementation of sort.Interface so that we can use built-in sort function (sort.Sort) for sorting SelectableData

// Len returns the length of data inside SelectableData.
func (d DataSelector) Len() int { return len(d.GenericDataList) }

// Swap swaps 2 indices inside SelectableData.
func (d DataSelector) Swap(i, j int) {
	d.GenericDataList[i], d.GenericDataList[j] = d.GenericDataList[j], d.GenericDataList[i]
}

// Less compares 2 indices inside SelectableData and returns true if first index is larger.
func (d DataSelector) Less(i, j int) bool {
	for _, sortBy := range d.Query.SortQuery.SortByList {
		a := d.GenericDataList[i].GetProperty(sortBy.Property)
		b := d.GenericDataList[j].GetProperty(sortBy.Property)
		// ignore sort completely if property name not found
		if a == nil || b == nil {
			break
		}
		cmp := a.Compare(b)
		if cmp == 0 { // values are the same. Just continue to next sortBy
			continue
		}
		return (cmp == -1 && sortBy.Ascending) || (cmp == 1 && !sortBy.Ascending)
	}
	return false
}

// Sort sorts the data inside as instructed by Query and returns itself to allow method chaining.
func (d *DataSelector) Sort() *DataSelector {
	sort.Sort(*d)
	return d
}

// Filter the data inside as instructed by Query and returns itself to allow method chaining.
func (d *DataSelector) Filter() *DataSelector {
	filteredList := []DataCell{}

	for _, c := range d.GenericDataList {
		matches := true
		for _, filterBy := range d.Query.FilterQuery.FilterByList {
			v := c.GetProperty(filterBy.Property)
			if v == nil || !v.Contains(filterBy.Value) {
				matches = false
				break
			}
		}
		if matches {
			filteredList = append(filteredList, c)
		}
	}

	d.GenericDataList = filteredList
	return d
}

// Paginate the data inside as instructed by Query and returns itself to allow method chaining.
func (d *DataSelector) Paginate() *DataSelector {
	pQuery := d.Query.PaginationQuery
	dataList := d.GenericDataList
	startIndex, endIndex := pQuery.GetPaginationSettings(len(dataList))

	// Return all items if provided settings do not meet requirements
	if !pQuery.IsValidPagination() {
		return d
	}
	// Return no items if requested page does not exist
	if !pQuery.IsPageAvailable(len(d.GenericDataList), startIndex) {
		d.GenericDataList = []DataCell{}
		return d
	}

	d.GenericDataList = dataList[startIndex:endIndex]
	return d
}

// GenericDataSelect takes a list of GenericDataCells and Query and returns selected data as instructed by dsQuery.
func GenericDataSelect(dataList []DataCell, dsQuery *Query) []DataCell {
	SelectableData := DataSelector{
		GenericDataList: dataList,
		Query:           dsQuery,
	}
	return SelectableData.Sort().Paginate().GenericDataList
}

// GenericDataSelectWithFilter takes a list of GenericDataCells and Query and returns selected data as instructed by dsQuery.
func GenericDataSelectWithFilter(dataList []DataCell, dsQuery *Query) ([]DataCell, int) {
	SelectableData := DataSelector{
		GenericDataList: dataList,
		Query:           dsQuery,
	}
	// Pipeline is Filter -> Sort -> CollectMetrics -> Paginate
	filtered := SelectableData.Filter()
	filteredTotal := len(filtered.GenericDataList)
	processed := filtered.Sort().Paginate()
	return processed.GenericDataList, filteredTotal
}

// GetComparableLabelFromMap returns a comprable label from a map[string]string
func GetComparableLabelFromMap(data map[string]string) ComparableValue {
	values := []string{}
	for k, v := range data {
		values = append(values, k+":"+v)
	}
	return StdComparableLabel(strings.Join(values, ","))
}
