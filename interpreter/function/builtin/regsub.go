// Code generated by __generator__/interpreter.go at once

package builtin

import (
	"regexp"

	"github.com/ysugimoto/falco/interpreter/context"
	"github.com/ysugimoto/falco/interpreter/function/errors"
	"github.com/ysugimoto/falco/interpreter/value"
)

const Regsub_Name = "regsub"

var Regsub_ArgumentTypes = []value.Type{value.StringType, value.StringType, value.StringType}

func Regsub_Validate(args []value.Value) error {
	if len(args) != 3 {
		return errors.ArgumentNotEnough(Regsub_Name, 3, args)
	}
	for i := range args {
		if args[i].Type() != Regsub_ArgumentTypes[i] {
			if args[i].Type() == value.BackendType && Regsub_ArgumentTypes[i] == value.StringType {
				v := args[i].(*value.Backend).Value.Name.Value
				args[i] = &value.String{Value: v}
			} else {
				return errors.TypeMismatch(Regsub_Name, i+1, Regsub_ArgumentTypes[i], args[i].Type())
			}
		}
	}
	return nil
}

func convertGoExpandString(replacement string) (string, bool) {
	var converted []rune
	var found bool
	repl := []rune(replacement)

	for i := 0; i < len(repl); i++ {
		r := repl[i]
		if r != 0x5C { // escape sequence, "\"
			converted = append(converted, r)
			continue
		}
		// If rune is escape sequence, find next numeric character which indicates matched index like "\1"
		var matchIndex []rune
		for {
			if i+1 > len(repl)-1 {
				break
			}
			r = repl[i+1]
			if r >= 0x31 && r <= 0x39 {
				matchIndex = append(matchIndex, r)
				i++
				continue
			}
			break
		}
		if len(matchIndex) > 0 {
			converted = append(converted, []rune("${"+string(matchIndex)+"}")...)
			found = true
		}
	}

	return string(converted), found
}

// Fastly built-in function implementation of regsub
// Arguments may be:
// - STRING, STRING, STRING
// Reference: https://developer.fastly.com/reference/vcl/functions/strings/regsub/
func Regsub(ctx *context.Context, args ...value.Value) (value.Value, error) {
	// Argument validations
	if err := Regsub_Validate(args); err != nil {
		return value.Null, err
	}

	input := value.Unwrap[*value.String](args[0])
	pattern := value.Unwrap[*value.String](args[1])
	replacement := value.Unwrap[*value.String](args[2])

	re, err := regexp.Compile(pattern.Value)
	if err != nil {
		ctx.FastlyError = &value.String{Value: "EREGRECUR"}
		return &value.String{Value: input.Value}, errors.New(
			Regsub_Name, "Invalid regular expression pattern: %s", pattern.Value,
		)
	}

	// Note: VCL's regsub uses PCRE regexp but golang is not PCRE
	matches := re.FindStringSubmatchIndex(input.Value)
	if matches == nil {
		return &value.String{Value: input.Value}, nil
	}

	if expand, found := convertGoExpandString(replacement.Value); found {
		replaced := re.ExpandString([]byte{}, expand, input.Value, matches)
		return &value.String{Value: string(replaced)}, nil
	}
	var replaced string
	if matches[0] > 0 {
		replaced += input.Value[:matches[0]]
	}
	replaced += replacement.Value
	if matches[1] < len(input.Value)-1 {
		replaced += input.Value[matches[1]:]
	}
	return &value.String{Value: replaced}, nil
}
