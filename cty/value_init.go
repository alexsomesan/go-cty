package cty

import (
	"fmt"
	"math/big"

	"golang.org/x/text/unicode/norm"

	"github.com/apparentlymart/go-cty/cty/set"
)

// BoolVal returns a Value of type Number whose internal value is the given
// bool.
func BoolVal(v bool) Value {
	return Value{
		ty: Bool,
		v:  v,
	}
}

// NumberVal returns a Value of type Number whose internal value is the given
// big.Float. The returned value becomes the owner of the big.Float object,
// and so it's forbidden for the caller to mutate the object after it's
// wrapped in this way.
func NumberVal(v *big.Float) Value {
	return Value{
		ty: Number,
		v:  v,
	}
}

// NumberIntVal returns a Value of type Number whose internal value is equal
// to the given integer.
func NumberIntVal(v int64) Value {
	return NumberVal(new(big.Float).SetInt64(v))
}

// NumberFloatVal returns a Value of type Number whose internal value is
// equal to the given float.
func NumberFloatVal(v float64) Value {
	return NumberVal(new(big.Float).SetFloat64(v))
}

// StringVal returns a Value of type String whose internal value is the
// given string.
//
// Strings must be UTF-8 encoded sequences of valid unicode codepoints, and
// they are NFC-normalized on entry into the world of cty values.
//
// If the given string is not valid UTF-8 then behavior of string operations
// is undefined.
func StringVal(v string) Value {
	return Value{
		ty: String,
		v:  norm.NFC.String(v),
	}
}

// ObjectVal returns a Value of an object type whose structure is defined
// by the key names and value types in the given map.
func ObjectVal(attrs map[string]Value) Value {
	attrTypes := make(map[string]Type, len(attrs))
	attrVals := make(map[string]interface{}, len(attrs))

	for attr, val := range attrs {
		attrTypes[attr] = val.ty
		attrVals[attr] = val.v
	}

	return Value{
		ty: Object(attrTypes),
		v:  attrVals,
	}
}

// ListVal returns a Value of list type whose element type is defined by
// the types of the given values, which must be homogenous.
//
// If the types are not all consistent (aside from elements that are of the
// dynamic pseudo-type) then this function will panic. It will panic also
// if the given list is empty, since then the element type cannot be inferred.
// (See also ListValEmpty.)
func ListVal(vals []Value) Value {
	if len(vals) == 0 {
		panic("must not call ListVal with empty slice")
	}
	elementType := DynamicPseudoType
	rawList := make([]interface{}, len(vals))

	for i, val := range vals {
		if elementType == DynamicPseudoType {
			elementType = val.ty
		} else if val.ty != DynamicPseudoType && !elementType.Equals(val.ty) {
			panic(fmt.Errorf(
				"inconsistent list element types (%#v then %#v)",
				elementType, val.ty,
			))
		}

		rawList[i] = val.v
	}

	return Value{
		ty: List(elementType),
		v:  rawList,
	}
}

// ListValEmpty returns an empty list of the given element type.
func ListValEmpty(element Type) Value {
	return Value{
		ty: List(element),
		v:  []interface{}{},
	}
}

// MapVal returns a Value of a map type whose element type is defined by
// the types of the given values, which must be homogenous.
//
// If the types are not all consistent (aside from elements that are of the
// dynamic pseudo-type) then this function will panic. It will panic also
// if the given map is empty, since then the element type cannot be inferred.
// (See also MapValEmpty.)
func MapVal(vals map[string]Value) Value {
	if len(vals) == 0 {
		panic("must not call MapVal with empty map")
	}
	elementType := DynamicPseudoType
	rawMap := make(map[string]interface{}, len(vals))

	for key, val := range vals {
		if elementType == DynamicPseudoType {
			elementType = val.ty
		} else if val.ty != DynamicPseudoType && !elementType.Equals(val.ty) {
			panic(fmt.Errorf(
				"inconsistent map element types (%#v then %#v)",
				elementType, val.ty,
			))
		}

		rawMap[key] = val.v
	}

	return Value{
		ty: Map(elementType),
		v:  rawMap,
	}
}

// MapValEmpty returns an empty map of the given element type.
func MapValEmpty(element Type) Value {
	return Value{
		ty: Map(element),
		v:  map[string]interface{}{},
	}
}

// SetVal returns a Value of set type whose element type is defined by
// the types of the given values, which must be homogenous.
//
// If the types are not all consistent (aside from elements that are of the
// dynamic pseudo-type) then this function will panic. It will panic also
// if the given list is empty, since then the element type cannot be inferred.
// (See also SetValEmpty.)
func SetVal(vals []Value) Value {
	if len(vals) == 0 {
		panic("must not call SetVal with empty slice")
	}
	elementType := DynamicPseudoType
	rawList := make([]interface{}, len(vals))

	for i, val := range vals {
		if elementType == DynamicPseudoType {
			elementType = val.ty
		} else if val.ty != DynamicPseudoType && !elementType.Equals(val.ty) {
			panic(fmt.Errorf(
				"inconsistent set element types (%#v then %#v)",
				elementType, val.ty,
			))
		}

		rawList[i] = val.v
	}

	rawVal := set.NewSetFromSlice(setRules{elementType}, rawList)

	return Value{
		ty: Set(elementType),
		v:  rawVal,
	}
}

// SetValEmpty returns an empty set of the given element type.
func SetValEmpty(element Type) Value {
	return Value{
		ty: Set(element),
		v:  set.NewSet(setRules{element}),
	}
}