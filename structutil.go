package structutil

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/golang/glog"
)

const conflictingValMsgPrefix = "Can't override conflicting values"

type ConflictPolicy struct {
	Override                bool
	TolerateConflictChecker func(key interface{}) bool
}

func (entity ConflictPolicy) TolerateConflict(key interface{}) bool {
	if entity.TolerateConflictChecker == nil {
		return true
	}
	return entity.TolerateConflictChecker(key)
}

func isSimpleType(val interface{}) bool {
	rt := reflect.TypeOf(val)
	switch rt.Kind() {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint,
	reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr, reflect.Float32, reflect.Float64,
	reflect.Complex64, reflect.Complex128, reflect.String:
		return true
	}
	return false
}

func isSliceStr(val interface{}) bool {
	if _, ok := val.([]string); ok {
		return true
	}
	return false
}

func isMapIntf(val interface{}) bool {
	if _, ok := val.(map[interface{}]interface{}); ok {
		return true
	}
	return false
}

func validateVal(val interface{}) error {
	if isSimpleType(val) || isMapIntf(val) || isSliceStr(val) {
		return nil
	}
	return errors.New(fmt.Sprintf("Encountered unsupported type %v\n" +
		"Supported types: simple | []string | map[interface{}]interface{}", reflect.TypeOf(val)))
}

func Stringify(in map[interface{}]interface{}, special map[interface{}]func(interface{}) string) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range in {
		if subMap, ok := v.(map[interface{}]interface{}); ok {
			result[fmt.Sprintf("%v", k)] = Stringify(subMap, special)
		} else {
			set := false
			if special != nil {
				if specialFun, ok := special[k]; ok {
					set = true
					result[fmt.Sprintf("%v", k)] = specialFun(v)
				}
			}
			if !set {
				result[fmt.Sprintf("%v", k)] = fmt.Sprintf("%v", v)
			}
		}
	}

	return result
}

func MakeMarshalFriendly(in map[interface{}]interface{}) *map[string]interface{} {
	result := make(map[string]interface{})
	for key, value := range in {
		keyAsStr := fmt.Sprintf("%v", key)
		valueAsMap, ok := value.(map[interface{}]interface{})
		if ok {
			result[keyAsStr] = *MakeMarshalFriendly(valueAsMap)
		} else {
			result[keyAsStr] = value
		}
	}

	return &result
}

func CloneMap(amap map[interface{}]interface{},
	typeSampleValues ...interface{}) map[interface{}]interface{} {
	if typeSampleValues != nil {
		for _, value := range typeSampleValues {
			gob.Register(value)
		}
	}
	buff := new(bytes.Buffer)
	enc := gob.NewEncoder(buff)
	dec := gob.NewDecoder(buff)
	err := enc.Encode(amap)
	if err != nil {
		panic(err)
	}
	res := make(map[interface{}]interface{})
	err = dec.Decode(&res)
	if err != nil {
		panic(err)
	}
	return res
}

func sliceRepr(in []interface{}, sep string) string {
	asStr := make([]string, len(in))
	for i, el := range in {
		asStr[i] = fmt.Sprintf("%v", el)
	}
	return strings.Join(asStr, sep)
}

func Merge(
	one map[interface{}]interface{}, // target
	two map[interface{}]interface{}, // source
	immutable bool, conflictPolicy ConflictPolicy,
	typeSampleValues ...interface{}) (map[interface{}]interface{}, error) {

	var source map[interface{}]interface{}
	var target map[interface{}]interface{}
	if immutable {
		source = CloneMap(two, typeSampleValues...)
		target = CloneMap(one, typeSampleValues...)
	} else {
		source = two
		target = one
	}
	var inner func(
		source map[interface{}]interface{},
		target map[interface{}]interface{},
		priorKeys []interface{}) (map[interface{}]interface{}, error)
	inner = func(
		source map[interface{}]interface{},
		target map[interface{}]interface{},
		priorKeys []interface{}) (map[interface{}]interface{}, error) {

		for sourceKey, sourceValue := range source {
			if err := validateVal(sourceValue); err != nil {
				return nil, err
			}
			keysRepr := func() string {
				priorKeysAsDn := sliceRepr(priorKeys, "/")
				if len(priorKeysAsDn) > 0 {
					priorKeysAsDn = "/" + priorKeysAsDn
				}
				return priorKeysAsDn + "/" + fmt.Sprintf("%v", sourceKey)
			}
			targetValue, existsInTarget := target[sourceKey]
			if existsInTarget {
				if sourceValue != nil && targetValue != nil &&
					reflect.TypeOf(targetValue) != reflect.TypeOf(sourceValue) {
					err := fmt.Sprintf(
						"different types at key %s detected in trying to merge maps",
						keysRepr())
					glog.Error(err)
					return nil, errors.New(err)
				}
				// no need to validate type because it's checked by combination of source type validation
				// and src/dst type equality check.

				if (targetValue == nil || isSimpleType(targetValue)) && sourceValue != targetValue {
					if conflictPolicy.TolerateConflict(sourceKey) {
						if conflictPolicy.Override {
							target[sourceKey] = sourceValue
						}
					} else {
						errMsg := fmt.Sprintf(
							"%s %s, %s at key %s", conflictingValMsgPrefix,
							fmt.Sprintf("%v", sourceValue), fmt.Sprintf("%v", targetValue), keysRepr())
						glog.Error(errMsg)
						return nil, errors.New(errMsg)
					}
				} else if isMapIntf(targetValue){
					if _, err := inner(sourceValue.(map[interface{}]interface{}), targetValue.(map[interface{}]interface{}), append(priorKeys, sourceKey)); err != nil {
						return nil, err
					}
				} else if isSliceStr(targetValue) {
					// put into target values that are in source but not yet in target
					s := sourceValue.([]string)
					t := targetValue.([]string)
					tMap := make(map[string]bool)
					for _, tv := range t {
						tMap[tv] = true
					}
					for _, sv := range s {
						if _, ok := tMap[sv]; !ok {
							t = append(t, sv)
						}
					}
					target[sourceKey] = t
				}
			} else {
				target[sourceKey] = source[sourceKey]
			}
		}

		return target, nil
	}
	return inner(source, target, make([]interface{}, 0))
}

func MergeAll(ms []map[interface{}]interface{}, immutable bool, conflictPolicy ConflictPolicy,
	typeSampleValues ...interface{}) (map[interface{}]interface{}, error) {
	type MergeResult = struct {
		result map[interface{}]interface{}
		err    error
	}
	switch size := len(ms); size {
	case 0:
		return nil, errors.New("mergeall on 0 elements? really?")
	case 1:
		return ms[0], nil
	default:
		mid := int(size / 2)
		chOne := make(chan MergeResult)
		chTwo := make(chan MergeResult)
		go func() {
			res, err := MergeAll(ms[0:mid], immutable, conflictPolicy, typeSampleValues...)
			chOne <- MergeResult{res, err}
		}()
		go func() {
			res, err := MergeAll(ms[mid:], immutable, conflictPolicy, typeSampleValues...)
			chTwo <- MergeResult{res, err}
		}()
		one := <-chOne
		if one.err != nil {
			return nil, one.err
		}
		two := <-chTwo
		if two.err != nil {
			return nil, two.err
		}
		return Merge(one.result, two.result, immutable, conflictPolicy, typeSampleValues...)
	}
}

// depthLimit = 0 for no depth limit
func FindMatching(dict map[interface{}]interface{}, matcher func(map[interface{}]interface{}) bool,
	maxdept int, stoponsuccess bool) []map[interface{}]interface{} {
	result := make([]map[interface{}]interface{}, 0)
	stop := false
	if matcher(dict) {
		result = append(result, dict)
		if stoponsuccess {
			stop = true
		}
	}
	if maxdept == 1 {
		stop = true
	}
	if !stop {
		for _, v := range dict {
			if childDict, ok := v.(map[interface{}]interface{}); ok {
				result = append(result, FindMatching(childDict, matcher, maxdept-1, stoponsuccess)...)
			}
		}
	}

	return result
}

func VisitMatching(root map[interface{}]interface{},
	matcher func(map[interface{}]interface{}) bool,
	nodeVisitor func(map[interface{}]interface{}) (interface{}, error),
	reducer func([]interface{}) (interface{}, error),
	maxdept int, stoponsuccess bool) (interface{}, error) {
	var results []interface{}
	for _, node := range FindMatching(root, matcher, maxdept, stoponsuccess) {
		interimResult, err := nodeVisitor(node)
		if err != nil {
			return nil, err
		}
		results = append(results, interimResult)
	}
	var result interface{}
	result = nil
	if reducer != nil {
		var err error
		result, err = reducer(results)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

func Contains(s []interface{}, e interface{}) bool {
	for _, el := range s {
		if el == e {
			return true
		}
	}
	return false
}

func FindInSlice(s []interface{}, e interface{}, eq func(interface{}, interface{}) bool) interface{} {
	for _, el := range s {
		if eq(el, e) {
			return el
		}
	}
	return nil
}

func DeleteIndexFromSlice(s []interface{}, i int) {
	copy(s[i:], s[i+1:])
	s[len(s)-1] = nil
	s = s[:len(s)-1]
}

func DeleteElFromSlice(s []interface{}, el interface{}) {
	index := -1
	for i, next := range s {
		if next == el {
			index = i
			break
		}
	}
	if index > -1 {
		DeleteIndexFromSlice(s, index)
	}
}

type OrderedSet struct {
	List []interface{}
	Map  map[interface{}]struct{}
}

func (s *OrderedSet) append(el interface{}) bool {
	if _, ok := s.Map[el]; !ok {
		s.Map[el] = struct{}{}
		s.List = append(s.List, el)
		return true
	} else {
		return false
	}
}

func (s *OrderedSet) delete(el interface{}) bool {
	if _, ok := s.Map[el]; !ok {
		return false
	} else {
		delete(s.Map, el)
		DeleteElFromSlice(s.List, el)
		return true
	}
}

func (s *OrderedSet) len() int {
	return len(s.List)
}

func ListToOrderedSet(l []interface{}) OrderedSet {
	var result OrderedSet
	for _, el := range l {
		result.append(el)
	}
	return result
}
