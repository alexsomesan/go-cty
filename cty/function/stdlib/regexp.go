package stdlib

import (
	"fmt"
	"regexp"
	resyntax "regexp/syntax"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

var RegexFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name: "pattern",
			Type: cty.String,
		},
		{
			Name: "string",
			Type: cty.String,
		},
	},
	Type: func(args []cty.Value) (cty.Type, error) {
		if !args[0].IsKnown() {
			// We can't predict our type without seeing our pattern
			return cty.DynamicPseudoType, nil
		}

		retTy, err := regexPatternResultType(args[0].AsString())
		if err != nil {
			err = function.NewArgError(0, err)
		}
		return retTy, err
	},
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		if retType == cty.DynamicPseudoType {
			return cty.DynamicVal, nil
		}

		re, err := regexp.Compile(args[0].AsString())
		if err != nil {
			// Should never happen, since we checked this in the Type function above.
			return cty.NilVal, function.NewArgErrorf(0, "error parsing pattern: %s", err)
		}
		str := args[1].AsString()

		captureIdxs := re.FindStringSubmatchIndex(str)
		if captureIdxs == nil {
			return cty.NilVal, fmt.Errorf("pattern did not match any part of the given string")
		}

		switch {
		case retType == cty.String:
			start, end := captureIdxs[0], captureIdxs[1]
			return cty.StringVal(str[start:end]), nil
		case retType.IsTupleType():
			captureIdxs = captureIdxs[2:] // index 0 is the whole pattern span, which we ignore by skipping one pair
			vals := make([]cty.Value, len(captureIdxs)/2)
			for i := range vals {
				start, end := captureIdxs[i*2], captureIdxs[i*2+1]
				if start < 0 || end < 0 {
					vals[i] = cty.NullVal(cty.String) // Did not match anything because containing group didn't match
					continue
				}
				vals[i] = cty.StringVal(str[start:end])
			}
			return cty.TupleVal(vals), nil
		case retType.IsObjectType():
			captureIdxs = captureIdxs[2:] // index 0 is the whole pattern span, which we ignore by skipping one pair
			vals := make(map[string]cty.Value, len(captureIdxs)/2)
			names := re.SubexpNames()[1:]
			for i, name := range names {
				start, end := captureIdxs[i*2], captureIdxs[i*2+1]
				if start < 0 || end < 0 {
					vals[name] = cty.NullVal(cty.String) // Did not match anything because containing group didn't match
					continue
				}
				vals[name] = cty.StringVal(str[start:end])
			}
			return cty.ObjectVal(vals), nil
		default:
			// Should never happen
			return cty.NilVal, fmt.Errorf("invalid return type: %s", retType.FriendlyNameForConstraint())
		}
	},
})

// Regex is a function that extracts one or more substrings from a given
// string by applying a regular expression pattern.
//
// The return type depends on the composition of the capture groups (if any)
// in the pattern:
//
//   - If there are no capture groups at all, the result is a single string
//     representing the entire matched pattern.
//   - If all of the capture groups are named, the result is an object whose
//     keys are the named groups and whose values are their sub-matches, or
//     null if a particular sub-group was inside another group that didn't
//     match.
//   - If all of the capture groups are named, the result is a tuple whose
//     elements are the sub-groups in order and whose values are their
//     sub-matches, or null if a particular sub-group was inside another group
//     that didn't match.
//   - It is invalid to use both named and un-named capture groups together in
//     the same pattern.
func Regex(pattern, str cty.Value) (cty.Value, error) {
	return RegexFunc.Call([]cty.Value{pattern, str})
}

// regexPatternResultType parses the given regular expression pattern and
// returns the structural type that would be returned to represent its
// capture groups.
//
// Returns an error if parsing fails or if the pattern uses a mixture of
// named and unnamed capture groups, which is not permitted.
func regexPatternResultType(pattern string) (cty.Type, error) {
	re, rawErr := regexp.Compile(pattern)
	switch err := rawErr.(type) {
	case *resyntax.Error:
		return cty.NilType, fmt.Errorf("invalid regexp pattern: %s in %s", err.Code, err.Expr)
	case error:
		// Should never happen, since all regexp compile errors should
		// be resyntax.Error, but just in case...
		return cty.NilType, fmt.Errorf("error parsing pattern: %s", err)
	}

	allNames := re.SubexpNames()[1:]
	var names []string
	unnamed := 0
	for _, name := range allNames {
		if name == "" {
			unnamed++
		} else {
			if names == nil {
				names = make([]string, 0, len(allNames))
			}
			names = append(names, name)
		}
	}
	switch {
	case unnamed == 0 && len(names) == 0:
		// If there are no capture groups at all then we'll return just a
		// single string for the whole match.
		return cty.String, nil
	case unnamed > 0 && len(names) > 0:
		return cty.NilType, fmt.Errorf("invalid regexp pattern: cannot mix both named and unnamed capture groups")
	case unnamed > 0:
		// For unnamed captures, we return a tuple of them all in order.
		etys := make([]cty.Type, unnamed)
		for i := range etys {
			etys[i] = cty.String
		}
		return cty.Tuple(etys), nil
	default:
		// For named captures, we return an object using the capture names
		// as keys.
		atys := make(map[string]cty.Type, len(names))
		for _, name := range names {
			atys[name] = cty.String
		}
		return cty.Object(atys), nil
	}
}