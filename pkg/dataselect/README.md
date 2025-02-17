# gomod.alauda.cn/alauda-backend/pkg/dataselect

This package offer a solution for Filter, Sorting and Pagging. 
Most of its code was copied from github.com/kubernetes/dashboard in order to make it easier to 
transition to this package. In a near future a better approach will be provided to solve filtering, sorting and pagging.

The most important parts of this package are listed bellow:

## Query (previously DataQuery)

A simple struct that holds Filter, Sort and Pagging options

## DataSelector

Besides holding the `Query` object, also holds an slice of `DataCell` in order
to execute the above mentioned operations in the desired order.

The `GenericDataSelectWithFilter(dataList []DataCell, dsQuery *Query) (result []DataCell, totalCount int)` function can be use to create DataSelector and filter data. `totalCount` will return the total number of items after filtering

## DataCell

In order to execute Filter and Sort it is necessary to provide a `DataCell` implenetation. The interface is defined as bellow

```
// DataCell describes the interface of the data cell that contains all the necessary methods needed to perform
// complex data selection
// GenericDataSelect takes a list of these interfaces and performs selection operation.
// Therefore as long as the list is composed of GenericDataCells you can perform any data selection!
type DataCell interface {
	// GetPropertyAtIndex returns the property of this data cell.
	// Value returned has to have Compare method which is required by Sort functionality of DataSelect.
	GetProperty(PropertyName) ComparableValue
}
```

## ComparableValue

Interface to compare and sort a specific value. i.e if `name` is given in the sort or filter query, the name will be encapsulated in a `ComparableValue`.


### Implementations

```
// StdComparableInt must be equal ints
type StdComparableInt int

// StdEqualString strings that must be exactly the same
type StdEqualString string

// StdComparableString only equal decoded UTF-8 values will return true
type StdComparableString string

// StdCaseInSensitiveComparableString case insensitive wrapper of StdComparableString
type StdCaseInSensitiveComparableString string

// StdComparableRFC3339Timestamp takes RFC3339 Timestamp strings and compares them as TIMES. In case of time parsing error compares values as strings.
type StdComparableRFC3339Timestamp string

// StdComparableTime time.Time implementation for ComperableValue
type StdComparableTime time.Time

// StdExactString exact comparisson of strings using == operator for ComparableValue
type StdExactString string

// StdComparableLabel label implementation of ComparableValue.
// supports multiple values split by comma ","
type StdComparableLabel string

// StdComparableStringIn comparable string in.
// Supports multiple values using ":" as a separator
// if any of the values is equal returns true
type StdComparableStringIn string

// StdComparableContainsString if one string contains the other
type StdComparableContainsString string
```

## ObjectDataCell

In order to simplify usage of filter, sort and pagging, a standard DataCell is offered for data inside the `metadata` or `ObjectMeta`. The implementation offers the following properties:

 - `name` as a `StdComparableContainsString`
 - `namespace` as a `StdComparableString`
 - `creationTimestamp` as a `StdComparableTime` (only for sorting)
 - `labels` as a `StdComparableLabel`
 - `annotations` as a `StdComparableLabel`

### Usage

Can be used for any `slice` of items that implement `metav1.Object` interface:

#### Covertions to and from []DataCell

The method `ToObjectCellSlice(slice interface{}) (values []DataCell)` can be used to convert a `slice` of items to `[]DataCell` and converted back to `[]metav1.Object` using the method `FromCellToObjectSlice(slice []DataCell) (values []metav1.Object)`.

Replacing the returned slice into the `ResourceList` object needs to be implemented seperatedly. Consult the `decorator` package for some helper methods.

