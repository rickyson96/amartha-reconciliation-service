package reconciliation

import (
	"maps"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestAppendMapOfSlices(t *testing.T) {
	type testMap map[string][]int
	tests := []struct {
		initial testMap
		key     string
		elem    int
		want    testMap
	}{
		{nil, "a", 1, testMap{"a": {1}}},
		{testMap{"a": {1}}, "a", 1, testMap{"a": {1, 1}}},
		{testMap{"a": {1, 1}}, "a", 1, testMap{"a": {1, 1, 1}}},
		{testMap{"a": {1, 1}}, "b", 1, testMap{"a": {1, 1}, "b": {1}}},
		{testMap{"a": {1}, "b": {1}}, "z", 2, testMap{"a": {1}, "b": {1}, "z": {2}}},
	}

	for _, test := range tests {
		init := maps.Clone(test.initial)
		got := appendMapOfSlices(test.initial, test.key, test.elem)
		if diff := cmp.Diff(test.want, got); diff != "" {
			t.Errorf("appendMapOfSlices(%v, %v, %v) mismatch, (-want, +got):\n%s", init, test.key, test.elem, diff)
		}
	}
}

func runTestAppend[M map[K][]V, K comparable, V any](t *testing.T, init M, key K, value V, want M) {
	t.Helper()
	initial := maps.Clone(init)
	got := appendMapOfSlices(init, key, value)
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("appendMapOfSlices(%v, %v, %v) mismatch, (-want, +got):\n%s", initial, key, value, diff)
	}
}

func TestAppendMapOfSlices_Types(t *testing.T) {
	runTestAppend(t, map[int][]int{1: {1}}, 1, 1, map[int][]int{1: {1, 1}})
	runTestAppend(t, map[string][]int{"a": {1}}, "a", 1, map[string][]int{"a": {1, 1}})
	runTestAppend(t, map[string][]string{"a": {"a"}}, "a", "a", map[string][]string{"a": {"a", "a"}})
	runTestAppend(t, map[bool][]bool{true: {true}}, true, true, map[bool][]bool{true: {true, true}})
	runTestAppend(t, map[float64][]float64{1.0: {1.0}}, 1.0, 1.0, map[float64][]float64{1.0: {1.0, 1.0}})
}

func TestPopMapOfSlices(t *testing.T) {
	type testMap map[string][]int
	tests := []struct {
		initial testMap
		key     string
		want    testMap
	}{
		{nil, "a", nil},
		{testMap{"a": {1}}, "a", testMap{}},
		{testMap{"a": {1, 1}}, "a", testMap{"a": {1}}},
		{testMap{"a": {1, 1}}, "b", testMap{"a": {1, 1}}},
		{testMap{"a": {1}, "b": {1}}, "b", testMap{"a": {1}}},
		{testMap{"a": {1, 1}, "b": {1}}, "a", testMap{"a": {1}, "b": {1}}},
	}

	for _, test := range tests {
		workingMap := maps.Clone(test.initial)
		popMapOfSlices(workingMap, test.key)
		if diff := cmp.Diff(test.want, workingMap); diff != "" {
			t.Errorf("popMapOfSlices(%v, %v) mismatch, (-want, +got):\n%s", test.initial, test.key, diff)
		}
	}
}

func runTestPop[M map[K][]V, K comparable, V any](t *testing.T, init M, key K, want M) {
	t.Helper()
	workingMap := maps.Clone(init)
	popMapOfSlices(workingMap, key)
	if diff := cmp.Diff(want, workingMap); diff != "" {
		t.Errorf("popMapOfSlices(%v, %v) mismatch, (-want, +got):\n%s", init, key, diff)
	}
}

func TestPopMapOfSlices_Types(t *testing.T) {
	runTestPop(t, map[int][]int{1: {1}}, 1, map[int][]int{})
	runTestPop(t, map[string][]int{"a": {1}}, "a", map[string][]int{})
	runTestPop(t, map[string][]string{"a": {"a"}}, "a", map[string][]string{})
	runTestPop(t, map[bool][]bool{true: {true}}, true, map[bool][]bool{})
	runTestPop(t, map[float64][]float64{1.0: {1.0}}, 1.0, map[float64][]float64{})
}
