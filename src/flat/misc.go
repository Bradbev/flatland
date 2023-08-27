package flat

func SliceToSet[T comparable](s []T) map[T]struct{} {
	ret := map[T]struct{}{}
	for _, elem := range s {
		ret[elem] = struct{}{}
	}
	return ret
}
