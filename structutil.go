package structutil

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"github.com/golang/glog"
	"strings"
)

func Stringify(in map[interface{}]interface{}, special map[interface{}] func(interface{}) string) map[string]interface{} {
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
	one map[interface{}]interface{},
	two map[interface{}]interface{},
	immutable bool,
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
			keysRepr := func() string {
				priorKeysAsDn := sliceRepr(priorKeys, "/")
				if len(priorKeysAsDn) > 0 {
					priorKeysAsDn = "/" + priorKeysAsDn
				}
				return priorKeysAsDn + "/" + fmt.Sprintf("%v", sourceKey)
			}
			targetValue, existsInTarget := target[sourceKey]
			if existsInTarget {
				targetValueAsMap, targetValueIsMap := targetValue.(map[interface{}]interface{})
				sourceValueAsMap, sourceValueIsMap := sourceValue.(map[interface{}]interface{})
				if targetValueIsMap != sourceValueIsMap {
					err := fmt.Sprintf(
						"different types at key %s detected in trying to merge maps",
						keysRepr())
					glog.Error(err)
					return nil, errors.New(err)
				} else if !targetValueIsMap {
					glog.Warning(fmt.Sprintf("Key clash at %s, value %v overrides value %v",
						keysRepr(), sourceValue, targetValue))
					target[sourceKey] = sourceValue
				} else {
					inner(sourceValueAsMap, targetValueAsMap, append(priorKeys, sourceKey))
				}

			} else {
				(target)[sourceKey] = (source)[sourceKey]
			}
		}

		return target, nil
	}
	return inner(source, target, make([]interface{}, 0))
}

func MergeAll(ms []map[interface{}]interface{}, immutable bool,
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
			res, err := MergeAll(ms[0:mid], immutable, typeSampleValues...)
			chOne <- MergeResult{res, err}
		}()
		go func() {
			res, err := MergeAll(ms[mid:], immutable, typeSampleValues...)
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
		return Merge(one.result, two.result, immutable, typeSampleValues...)
	}
}
