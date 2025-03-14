package reconciliation

// appendMapOfSlices appends to slices in a map.
// BEWARE: this behaves like usual append
// it will always modify current map, unless it's nil, which then
// it'll instantiate a new map to be returned.
// It's always recommended to use the following syntax:
//
//	m = appendMapOfSlices(m, key, elem)
func appendMapOfSlices[M ~map[K]V, K comparable, V ~[]E, E any](m M, key K, elem E) M {
	if m == nil {
		m = make(map[K]V)
		m[key] = []E{elem}
		return m
	}

	_, ok := m[key]
	if !ok {
		m[key] = []E{elem}
		return m
	}

	m[key] = append(m[key], elem)
	return m
}

// popMapOfSlices pops the first value out of the slice inside a map based on key
// it'll delete the key when it pops the last value
// if the key doesn't found, it doesn't alter the map
func popMapOfSlices[M ~map[K]S, K comparable, S ~[]E, E any](m M, key K) {
	if m == nil {
		return
	}

	existing, ok := m[key]
	if !ok {
		return
	}

	if len(existing) == 1 {
		delete(m, key)
		return
	}

	m[key] = m[key][1:]
}
