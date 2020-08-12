package structutil

import (
	"encoding/json"
	"gotest.tools/v3/assert"
	"testing"
)

var TypeValuesToRegister []interface{}

var one map[interface{}]interface{}
var two map[interface{}]interface{}
var twoPrime map[interface{}]interface{}
var three map[interface{}]interface{}
var four map[interface{}]interface{}
var five map[interface{}]interface{}
var six map[interface{}]interface{}
var mergeResult map[interface{}]interface{}
var intermixed map[interface{}]interface{}
var intermixedMergeResult map[interface{}]interface{}
var intermixedBad map[interface{}]interface{}

func initialize() {
	TypeValuesToRegister = []interface{}{map[interface{}]interface{}{}, map[string]interface{}{}, map[string]string{}}

	one = map[interface{}]interface{}{
		"keyOne1": map[interface{}]interface{}{
			"keyOne1_1": "valueOneOne1_1",
		},
		"keyOne2": "valueOne2",
		"keyOne3": map[interface{}]interface{}{
			"keyOne3_1": "valueOne3_1",
			"keyOne3_2": "valueOne3_2",
			"keyOne3_3": map[interface{}]interface{}{
				"keyOne4_1": "valueOne4_1",
			},
		},
	}
	two = map[interface{}]interface{}{
		"keyTwo1": map[interface{}]interface{}{
			"keyTwo1_1": "valueTwo1_1",
		},
		"keyTwo2": "valueTwo2",
		"keyTwo3": map[interface{}]interface{}{
			"keyTwo3_1": "valueTwo3_1",
			"keyTwo3_2": "valueTwo3_2",
			"keyTwo3_3": map[interface{}]interface{}{
				"keyTwo4_1": "valueTwo4_1",
			},
		},
		"keyTwo4": "valueTwo4",
	}
	twoPrime = map[interface{}]interface{}{
		"keyTwo1": map[interface{}]interface{}{
			"keyTwo1_1": "valueTwo1_1Prime",
		},
	}
	three = map[interface{}]interface{}{
		"keyThree1": map[interface{}]interface{}{
			"keyThree1_1": map[interface{}]interface{}{
				"keyThree1_1_1": "valueThree1_1_1",
			},
		},
		"keyThree2": "valueThree2",
		"keyThree3": map[interface{}]interface{}{
			"keyThree3_1": "valueThree3_1",
			"keyThree3_2": "valueThree3_2",
			"keyThree3_3": map[interface{}]interface{}{
				"keyThree4_1": "valueThree4_1",
			},
		},
	}
	four = map[interface{}]interface{}{
		"keyFour1": map[interface{}]interface{}{
			"keyFour1_1": map[interface{}]interface{}{
				"keyFour1_1_1": "valueFour1_1_1",
			},
		},
	}
	five = map[interface{}]interface{}{
		"keyFive1": "valueFive1",
	}
	six = map[interface{}]interface{}{
		"keySix1": "valueSix1",
	}

	mergeResult = map[interface{}]interface{}{
		"keyFive1": "valueFive1",
		"keyFour1": map[interface{}]interface{}{
			"keyFour1_1": map[interface{}]interface{}{
				"keyFour1_1_1": "valueFour1_1_1",
			},
		},
		"keyOne1": map[interface{}]interface{}{
			"keyOne1_1": "valueOneOne1_1",
		},
		"keyOne2": "valueOne2",
		"keyOne3": map[interface{}]interface{}{
			"keyOne3_1": "valueOne3_1",
			"keyOne3_2": "valueOne3_2",
			"keyOne3_3": map[interface{}]interface{}{
				"keyOne4_1": "valueOne4_1",
			},
		},
		"keySix1": "valueSix1",
		"keyThree1": map[interface{}]interface{}{
			"keyThree1_1": map[interface{}]interface{}{
				"keyThree1_1_1": "valueThree1_1_1",
			},
		},
		"keyThree2": "valueThree2",
		"keyThree3": map[interface{}]interface{}{
			"keyThree3_1": "valueThree3_1",
			"keyThree3_2": "valueThree3_2",
			"keyThree3_3": map[interface{}]interface{}{
				"keyThree4_1": "valueThree4_1",
			},
		},
		"keyTwo1": map[interface{}]interface{}{
			"keyTwo1_1": "valueTwo1_1",
		},
		"keyTwo2": "valueTwo2",
		"keyTwo3": map[interface{}]interface{}{
			"keyTwo3_1": "valueTwo3_1",
			"keyTwo3_2": "valueTwo3_2",
			"keyTwo3_3": map[interface{}]interface{}{
				"keyTwo4_1": "valueTwo4_1",
			},
		},
		"keyTwo4": "valueTwo4",
	}
	intermixed = map[interface{}]interface{}{
		"keyTwo1": map[interface{}]interface{}{
			"keyOne1_1": map[interface{}]interface{}{
				"keyTwo1_1_1": "valueTwo1_1_1",
			},
		},
		"keyTwo3": map[interface{}]interface{}{
			"keyTwo3_1": "valueTwo3_1_modified",
		},
	}
	intermixedMergeResult = map[interface{}]interface{}{
		"keyTwo1": map[interface{}]interface{}{
			"keyOne1_1": map[interface{}]interface{}{
				"keyTwo1_1_1": "valueTwo1_1_1",
			},
			"keyTwo1_1": "valueTwo1_1",
		},
		"keyTwo2": "valueTwo2",
		"keyTwo3": map[interface{}]interface{}{
			"keyTwo3_1": "valueTwo3_1_modified",
			"keyTwo3_2": "valueTwo3_2",
			"keyTwo3_3": map[interface{}]interface{}{
				"keyTwo4_1": "valueTwo4_1",
			},
		},
		"keyTwo4": "valueTwo4",
	}

	intermixedBad = map[interface{}]interface{}{
		"keyTwo1": "valueTwo1",
	}

}

func merge1(t *testing.T, immutable bool) {
	initialize()
	mergeableMaps := []map[interface{}]interface{}{one, two, three, four, five, six}
	result, err := MergeAll(mergeableMaps, immutable, TypeValuesToRegister...)
	if err != nil {
		t.Fatal("Error reported where should be none")
	}
	assert.DeepEqual(t, result, mergeResult)
}

func TestMerge1(t *testing.T) {
	merge1(t, false)
	merge1(t, true)
}

func merge2(t *testing.T, immutable bool) {
	initialize()
	// mergeableMaps = []*map[interface{}]interface{} {&one, &two, &three, &four, &five, &six, &intermixed}
	mergeableMaps := []map[interface{}]interface{}{two, intermixed}
	result, err := MergeAll(mergeableMaps, immutable, TypeValuesToRegister...)
	if err != nil {
		t.Fatal("Error reported where should be none")
	}
	assert.DeepEqual(t, result, intermixedMergeResult)
}

func TestMerge2(t *testing.T) {
	merge2(t, false)
	merge2(t, true)
}

func TestMutable(t *testing.T) {
	initialize()
	preMergeKeyTwo1_1 := two["keyTwo1"].(map[interface{}]interface{})["keyTwo1_1"]
	Merge(two, twoPrime, false, TypeValuesToRegister...)
	postMergeKeyTwo1_1 := two["keyTwo1"].(map[interface{}]interface{})["keyTwo1_1"]
	assert.Check(t, preMergeKeyTwo1_1 != postMergeKeyTwo1_1)
}

func TestImmutable(t *testing.T) {
	initialize()
	preMergeKeyTwo1_1 := two["keyTwo1"].(map[interface{}]interface{})["keyTwo1_1"]
	Merge(two, twoPrime, true, TypeValuesToRegister...)
	postMergeKeyTwo1_1 := two["keyTwo1"].(map[interface{}]interface{})["keyTwo1_1"]
	assert.Equal(t, preMergeKeyTwo1_1, postMergeKeyTwo1_1)
}

func merge3(t *testing.T, immutable bool) {
	initialize()
	// mergeableMaps = []*map[interface{}]interface{} {&one, &two, &three, &four, &five, &six, &intermixed}
	mergeableMaps := []map[interface{}]interface{}{two, intermixedBad}
	result, err := MergeAll(mergeableMaps, immutable, TypeValuesToRegister...)
	if err == nil {
		resultJsonRepr, _ := json.Marshal(MakeMarshalFriendly(result))
		t.Fatal("Error should've been reported but wasn't. Got this instead:\n" + string(resultJsonRepr))
	}
}

func TestMerge3(t *testing.T) {
	merge3(t, false)
	merge3(t, true)
}

func merge4(t *testing.T, immutable bool) {
	initialize()
	// mergeableMaps = []*map[interface{}]interface{} {&one, &two, &three, &four, &five, &six, &intermixed}
	mergeableMaps := []map[interface{}]interface{}{intermixedBad, two}
	result, err := MergeAll(mergeableMaps, false, TypeValuesToRegister...)
	if err == nil {
		resultJsonRepr, _ := json.Marshal(MakeMarshalFriendly(result))
		t.Fatal("Error should've been reported but wasn't. Got this instead:\n" + string(resultJsonRepr))
	}
}

func TestMerge4(t *testing.T) {
	merge4(t, false)
	merge4(t, true)
}
