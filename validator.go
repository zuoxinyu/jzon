package jzon

import (
	"regexp"
	"fmt"
	"strings"
)

const sample = `
{
	"str": "regexp: /*/"
	"num": "range: 100-200 | -1+2 | 1,2,3"
	"bol": "bool: true|false"
	"nul": "null: null"
}
`

const (
	COND_REGEXP = "regexp:"
	COND_RANGE  = "range:"
	COND_BOOL   = "bool:"
	COND_NULL   = "null:"
)

const (
	RANGE_CLOSE int = iota
	RANGE_EXCEPT
	RANGE_ARRAY
	RANGE_ONE
)

const (
	LEVEL_FETAL int = iota
	LEVEL_EMERGENCY
	LEVEL_EXCEPTION
	LEVEL_LOWEST
)

// Validator defines a validator to validate if a JSON value can be acceptable
type Validator struct {
	Type string			// Type indicates this is used for validating which type
	Level int 			// Level defines the error level
	Reg  regexp.Regexp  // Reg verifies strings
	Rng  Range 			// Rng verifies numbers
	Bol  bool			// Bol verifies bool
}

// Range defines a numeric range, if the `Type` is RANGE_CLOSE,
// the target number should match LowerBound <= target < UpperBound
// otherwise should match target < LowerBound || target > UpperBound
type Range struct {
	Type       int
	UpperBound int64
	LowerBound int64
	Array      []int64
}

// Compile parses string to fill a Validator
func (v *Validator) Compile(cond string) (err error) {
	var reg *regexp.Regexp
	var rng Range

	switch {
	case strings.HasPrefix(cond, COND_REGEXP) :
		reg, err = regexp.Compile(strings.TrimPrefix(cond, COND_REGEXP))
		if err != nil {
			return err
		}

		v.Type = COND_REGEXP
		v.Reg = *reg
		return  nil

	case strings.HasPrefix(cond, COND_RANGE):
		rangeStr := strings.TrimPrefix(cond, COND_RANGE)
		nparsed, err := fmt.Sscanf(rangeStr, "%d,%d", &rng.LowerBound, &rng.UpperBound)
		if err != nil {
			return err
		}

		if nparsed != 2 {
			return fmt.Errorf("expect 2 range numbers, but found %d", nparsed)
		}

		v.Type = COND_RANGE
		v.Rng = rng
		return nil

	case strings.HasPrefix(cond, COND_BOOL):
		boolStr := strings.TrimPrefix(cond, COND_BOOL)
		switch boolStr {
		case "true":
			v.Type = COND_BOOL
			v.Bol = true
		case "false":
			v.Type = COND_BOOL
			v.Bol = false
		case "both":
			v.Type = COND_BOOL
		default:
			return fmt.Errorf("expect `true` | `fasle` | `both` but found `%s`", boolStr)
		}

	case strings.HasPrefix(cond, COND_NULL):
		nullStr := strings.TrimPrefix(cond, COND_NULL)
		if nullStr == "null" {
			v.Type = COND_NULL
			return nil
		}

		return nil
	}

	return fmt.Errorf("expect a string with prefix `%s` | `%s` | `%s` | `%s` but found `%s`",
		COND_REGEXP, COND_BOOL, COND_NULL, COND_RANGE, cond)
}

// Validate verifies this node by another JSON which has a particular format,
// the given JSON should define the format of each field by a validator and
// an level number. If there were some fields can't pass the relying validator,
// the level numbers would give errors respectively
func (jz *Jzon) Validate(validator *Jzon) (ok bool, err error) {
	return
}

// Match judges whether the target number should be accepted
func (rng *Range) CanAccept(n int64) bool {
	switch rng.Type{
	case RANGE_CLOSE:
		return rng.LowerBound <= n && n < rng.UpperBound
	case RANGE_EXCEPT:
		return n < rng.LowerBound || rng.UpperBound < n
	case RANGE_ONE:
		return n == rng.LowerBound
	case RANGE_ARRAY:
		has := false
		for _, v := range rng.Array {
			if v == n {
				has = true
				break
			}
		}
		return has
	}
	return false
}

