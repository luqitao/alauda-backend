/*
import from "k8s.io/kubernetes/pkg/util/hash"
*/

package hash

import (
	"crypto/md5"
	"encoding/hex"
	"hash"

	"github.com/davecgh/go-spew/spew"
)

// DeepHashObject writes specified object to hash using the spew library
// which follows pointers and prints actual values of the nested objects
// ensuring the hash does not change when a pointer changes.
func DeepHashObject(hasher hash.Hash, objectToWrite interface{}) {
	hasher.Reset()
	printer := spew.ConfigState{
		Indent:         " ",
		SortKeys:       true,
		DisableMethods: true,
		SpewKeys:       true,
	}
	printer.Fprintf(hasher, "%#v", objectToWrite)
}

// HashToString gen hash string base on actual values of the nested objects.
func HashToString(obj interface{}) string {
	hasher := md5.New()
	DeepHashObject(hasher, obj)
	return hex.EncodeToString(hasher.Sum(nil)[0:])
}
