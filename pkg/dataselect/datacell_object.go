package dataselect

import (
	"encoding/json"
	"reflect"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// PropertyGetter is a getter to get property
type PropertyGetter struct {
	// F is the func that how to get value from an object
	// the obj will be a pointer of k8s runtime object
	F func(obj interface{}) ComparableValue
	// Name is the target property name that you want to get
	Name PropertyName
}

// ObjectDataCell standard implementation of DataCell for metav1.Object
type ObjectDataCell struct {
	metav1.Object

	// Getters indicates how to get property by name
	Getters []PropertyGetter
}

var _ DataCell = ObjectDataCell{}

const (
	objectName              = "name"
	objectNamespace         = "namespace"
	objectLabel             = "labels"
	objectAnnotation        = "annotations"
	objectCreationTimestamp = "creationTimestamp"
)

// GetProperty returns a comparablevalue for metav1.Object datacell
func (o ObjectDataCell) GetProperty(name PropertyName) ComparableValue {
	if o.Object == nil {
		return nil
	}
	switch name {
	case objectName:
		return StdComparableContainsString(o.GetName())
	case objectNamespace:
		return StdComparableString(o.GetNamespace())
	case objectCreationTimestamp:
		return StdComparableTime(o.GetCreationTimestamp().Time)
	case objectLabel:
		if len(o.GetLabels()) > 0 {
			return GetComparableLabelFromMap(o.GetLabels())
		}
	case objectAnnotation:
		if len(o.GetAnnotations()) > 0 {
			return GetComparableLabelFromMap(o.GetAnnotations())
		}
	}

	for _, getter := range o.Getters {
		if string(getter.Name) == string(name) {
			if getter.F == nil {
				return nil
			}
			return getter.F(o.Object)
		}
	}

	return nil
}

// ToObjectCellSlice converts slice objects to []ObjectDataCell
func ToObjectCellSlice(slice interface{}) (values []DataCell) {
	return ToObjectCellSliceX(slice)
}

// ToObjectCellSliceX is same as ToObjectCellSlice but extending easily,
// you do not need to define a new ObjectCell, just add extra PropertyGetters to  ObjectCell
//
// example:
//   var DisplayZHNameProperty PropertyGetter = PropertyGetter{
// 	  Name: "displayZhName",
// 	  F: func(obj interface{}) dataselect.ComparableValue {
// 		  template := obj.(*v1alpha1.PipelineTemplate)
// 		  return dataselect.StdComparableString(template.Annotations[v1alpha1.AnnotationsKeyDisplayName])
// 	  },
//   }
// ToObjectCellSliceX(items, DisplayZHNameProperty)
func ToObjectCellSliceX(slice interface{}, getters ...PropertyGetter) (values []DataCell) {
	rv := reflect.ValueOf(slice)
	if rv.Kind() != reflect.Slice || rv.Len() == 0 {
		values = []DataCell{}
		return
	}
	// using reflection to fetch the whole
	// interface as a slice and iterate over it
	values = make([]DataCell, 0, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		item := rv.Index(i)
		// Kubernetes list items are not pointers,
		// metav1.Object interface is implemented in *metav1.ObjectMeta
		// so we need to fetch its pointer using Addr() then convert to the interface
		if obj, ok := item.Addr().Interface().(metav1.Object); ok {
			values = append(values, ObjectDataCell{Object: obj, Getters: getters})
		}
	}
	return
}

// FromCellToObjectSlice convert back to metav1.Object
func FromCellToObjectSlice(slice []DataCell) (values []metav1.Object) {
	values = make([]metav1.Object, 0, len(slice))
	if len(slice) > 0 {
		for _, s := range slice {
			if obj, ok := s.(ObjectDataCell); ok {
				values = append(values, obj.Object)
			}
		}
	}
	return
}

// FromCellToUnstructuredSlice convert objlist to unstructured slice
func FromCellToUnstructuredSlice(slice []DataCell) (items []unstructured.Unstructured) {
	items = make([]unstructured.Unstructured, 0, len(slice))
	if len(slice) > 0 {
		for _, s := range slice {
			if obj, ok := s.(ObjectDataCell); ok {
				if unsObj, ok := unstructuredObject(obj.Object); ok {
					items = append(items, *unsObj)
				}
			}
		}
	}
	return
}

// object to unstructured
func unstructuredObject(obj metav1.Object) (*unstructured.Unstructured, bool) {
	bs, err := json.Marshal(obj)
	if err != nil {
		return nil, false
	}
	ret := &unstructured.Unstructured{}
	if err = ret.UnmarshalJSON(bs); err != nil {
		return nil, false
	}
	return ret, true
}
