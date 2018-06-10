package jzon

import (
    "fmt"
    "regexp"
    "strings"
)

const sample = `
{
    "str": "regexp: 0 /*/"
    "num": "range: 2 100-200 | -1+2 | 1,2,3"
    "bol": "bool: 1 true|false"
    "nul": "null: 2 null"
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
)

const (
    LEVEL_FETAL int = iota
    LEVEL_EMERGENCY
    LEVEL_EXCEPTION
    LEVEL_LOWEST
)

// Validator defines a validator to validate if a JSON value can be acceptable
type Validator func(*Jzon) bool

// Range defines a numeric range
// for RANGE_CLOSE:
// 		LowerBound <= target < UpperBound
// for RANGE_EXCEPT:
// 		LowerBound > target || target > UpperBound
// for RANGE_ARRAY:
// 		array.Contains(target)
type Range struct {
    Type       int
    UpperBound int64
    LowerBound int64
    Array      []int64
}

// CanAccept judges whether the target number should be accepted
func (rng *Range) CanAccept(n int64) bool {
    switch rng.Type{
    case RANGE_CLOSE:
        return rng.LowerBound <= n && n < rng.UpperBound
    case RANGE_EXCEPT:
        return n < rng.LowerBound || rng.UpperBound < n
    case RANGE_ARRAY:
        for _, v := range rng.Array {
            if v == n {
                return true
            }
        }
        return false
    }
    return false
}

func compileCondition(cond string) (Validator, error) {
    switch {
    case strings.HasPrefix(cond, COND_REGEXP) :
        return compileRegExp(cond)

    case strings.HasPrefix(cond, COND_RANGE):
        return compileRange(cond)

    case strings.HasPrefix(cond, COND_BOOL):
        return compileBool(cond)

    case strings.HasPrefix(cond, COND_NULL):
        return compileNull(cond)
    }

    return nil, fmt.Errorf("expect a string with prefix `%s` | `%s` | `%s` | `%s` but found `%s`",
        COND_REGEXP, COND_BOOL, COND_NULL, COND_RANGE, cond)
}

func compileRegExp(cond string) (Validator, error) {
    reg, err := regexp.Compile(strings.TrimPrefix(cond, COND_REGEXP))
    if err != nil {
        return nil, err
    }

    return func(jz *Jzon) bool {
        str, err := jz.String()
        if err != nil {
            return false
        }
        return reg.Match([]byte(str))
    }, nil

}

func compileRange(cond string) (Validator, error) {
    var rng Range
    rangeStr := strings.TrimPrefix(cond, COND_RANGE)
    nParsed, err := fmt.Sscanf(rangeStr, "%d,%d", &rng.LowerBound, &rng.UpperBound)
    if err != nil {
        return nil, err
    }

    if nParsed != 2 {
        return nil, fmt.Errorf("expect 2 range numbers, but found %d", nParsed)
    }

    return func(jz *Jzon) bool {
        var n int64
        if n, err = jz.Number(); err != nil {
            return false
        }
        return rng.CanAccept(n)
    }, nil
}

func compileBool(cond string) (Validator, error) {
    boolStr := strings.TrimPrefix(cond, COND_BOOL)
    var b bool
    if boolStr == "true" {
        b = true
    } else if boolStr == "false" {
        b = false
    } else if boolStr == "both" {
        return func(jz *Jzon) bool {
            return jz.Type == JzTypeBol
        }, nil
    } else {
        return nil, fmt.Errorf("expect `true` | `fasle` | `both` but found `%s`", boolStr)
    }

    return func(jz *Jzon) bool {
        v, err := jz.Bool()
        if err != nil {
            return false
        }
        return v == b
    }, nil
}

func compileNull(cond string) (Validator, error) {
    nullStr := strings.TrimPrefix(cond, COND_NULL)
    if nullStr == "null" {
        return func(jz *Jzon) bool {
            return jz.Type == JzTypeNul
        }, nil
    }
    return nil, expectString("null", []byte(nullStr))
}

// Validate verifies this node by another JSON which has a particular grammar,
// the given JSON should define the format of each field by a validator and a
// level number. If there were some fields can't pass the relying validator,
// the level numbers would give errors respectively
func (jz *Jzon) Validate(validator *Jzon) (ok bool, err error) {
    return
}

