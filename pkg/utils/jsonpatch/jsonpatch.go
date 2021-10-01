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

// Item specifies a patch operation for a string.
type Item struct {
	Op    Operation   `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

type Operation string

const (
	OperationAdd     Operation = "add"
	OperationReplace Operation = "replace"
	OperationRemove  Operation = "remove"
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
		tmp := strings.ReplaceAll(path, "~", "~0")
		tmp = strings.ReplaceAll(tmp, "/", "~1")
		transferred = append(transferred, tmp)
	}
	return SetPath(false, transferred...)
}
