/*
Copyright 2024 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package validators

import (
	"k8s.io/gengo/v2/parser/tags"
	"k8s.io/gengo/v2/types"
)

const (
	// libValidationPkg is the pkgpath to our "standard library" of validation
	// functions.
	libValidationPkg = "k8s.io/apimachinery/pkg/api/validate"
)

func getMemberByJSON(t *types.Type, jsonName string) *types.Member {
	for i := range t.Members {
		if jsonTag, ok := tags.LookupJSON(t.Members[i]); ok {
			if jsonTag.Name == jsonName {
				return &t.Members[i]
			}
		}
	}
	return nil
}

// isNilableType returns true if the argument type can be compared to nil.
func isNilableType(t *types.Type) bool {
	for t.Kind == types.Alias {
		t = t.Underlying
	}
	switch t.Kind {
	case types.Pointer, types.Map, types.Slice, types.Interface: // Note: Arrays are not nilable
		return true
	}
	return false
}

// nativeType returns the Go native type of the argument type, with any
// intermediate typedefs removed. Go itself already flattens typedefs, but this
// handles it in the unlikely event that we ever fix that.
//
// Examples:
// * Trivial:
//   - given `int`, returns `int`
//   - given `*int`, returns `*int`
//   - given `[]int`, returns `[]int`
//
// * Typedefs
//   - given `type X int; X`, returns `int`
//   - given `type X int; []X`, returns `[]X`
//
// * Typedefs and pointers:
//   - given `type X int; *X`, returns `*int`
//   - given `type X *int; *X`, returns `**int`
//   - given `type X []int; X`, returns `[]int`
//   - given `type X []int; *X`, returns `*[]int`
func nativeType(t *types.Type) *types.Type {
	ptrs := 0
	for {
		if t.Kind == types.Alias {
			t = t.Underlying
		} else if t.Kind == types.Pointer {
			ptrs++
			t = t.Elem
		} else {
			break
		}
	}
	for range ptrs {
		t = types.PointerTo(t)
	}
	return t
}

// nonPointer returns the value-type of a possibly pointer type. If type is not
// a pointer, it returns the input type.
func nonPointer(t *types.Type) *types.Type {
	for t.Kind == types.Pointer {
		t = t.Elem
	}
	return t
}

// rootTypeString returns a string representation of the relationship between
// src and dst types, for use in error messages.
func rootTypeString(src, dst *types.Type) string {
	if src == dst {
		return src.String()
	}
	return src.String() + " -> " + dst.String()
}
