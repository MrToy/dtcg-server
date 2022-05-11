package app

import "errors"

func ListPick[T comparable](src []T, num int) ([]T, []T, error) {
	if len(src) < num {
		return nil, nil, errors.New("not enough")
	}
	return src[len(src)-num:], src[:len(src)-num], nil
}

func ListRemoveAt[T comparable](list []T, index int) []T {
	return append(list[:index], list[index+1:]...)
}

func ListRemove[T comparable](list []T, item T) []T {
	index := ListFindIndex(list, func(a T) bool {
		return a == item
	})
	return ListRemoveAt(list, index)
}

func ListAppend[T comparable](src []T, dst []T) []T {
	return append(dst, src...)
}

func ListMove[T comparable](src []T, dist []T, num int) ([]T, []T, error) {
	if len(src) < num {
		return nil, nil, errors.New("not enough")
	}
	return src[:len(src)-num], append(dist, src[len(src)-num:]...), nil
}

func ListDifference[T comparable](a, b []T) []T {
	mb := make(map[T]struct{}, len(b))
	for _, x := range b {
		mb[x] = struct{}{}
	}
	var diff []T
	for _, x := range a {
		if _, found := mb[x]; !found {
			diff = append(diff, x)
		}
	}
	return diff
}

func ListFindIndex[T comparable](list []T, predicate func(a T) bool) int {
	for i := 0; i < len(list); i++ {
		if predicate(list[i]) {
			return i
		}
	}
	return -1
}
