package jsonpatch

import "strings"

// JSON Patch is a format for describing changes to a JSON document.
// It can be used to avoid sending a whole document when only a part has changed.
// When used in combination with the HTTP PATCH method,
// it allows partial updates for HTTP APIs in a standards compliant way.
// An example:
//
// {
//   "baz": "qux",
//   "foo": "bar"
// }
//
// The patch:
//
// [
//   { "op": "replace", "path": "/baz", "value": "boo" },
//   { "op": "add", "path": "/hello", "value": ["world"] },
//   { "op": "remove", "path": "/foo" }
// ]
//
// The result:
//
// {
//   "baz": "boo",
//   "hello": ["world"]
// }
//
// See more in http://jsonpatch.com/

// JsonPatchItem specifies a patch operation for a string.
type JsonPatchItem struct {
	Op    JsonPatchOperation `json:"op"`
	Path  string             `json:"path"`
	Value interface{}        `json:"value,omitempty"`
}

type JsonPatchOperation string

const (
	OperationAdd     JsonPatchOperation = "add"
	OperationReplace JsonPatchOperation = "replace"
	OperationRemove  JsonPatchOperation = "remove"
)

// SetPath returns the Path for JsonPatchItem
// user must specify whether needs to be transferred
// if needed, `~` and `/` are escaped with `~0` and `~1` respectively.
func SetPath(needTransfer bool, paths ...string) string {
	if !needTransfer {
		path := strings.Join(paths, "/")
		return "/" + path
	}
	transferred := []string{}
	for _, path := range paths {
		tmp := strings.Replace(path, "~", "~0", -1)
		tmp = strings.Replace(tmp, "/", "~1", -1)
		transferred = append(transferred, tmp)
	}
	return SetPath(false, transferred...)
}