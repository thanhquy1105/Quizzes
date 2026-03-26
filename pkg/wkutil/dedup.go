package wkutil

import (
	"sort"
)

func RemoveRepeatedElementGeneric[T comparable](arr []T) []T {
	if len(arr) == 0 {
		return arr
	}

	seen := make(map[T]bool, len(arr))
	result := make([]T, 0, len(arr))

	for _, item := range arr {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}

func RemoveRepeatedElementSorted[T comparable](arr []T) []T {
	if len(arr) <= 1 {
		return arr
	}

	writeIndex := 1
	for readIndex := 1; readIndex < len(arr); readIndex++ {
		if arr[readIndex] != arr[readIndex-1] {
			arr[writeIndex] = arr[readIndex]
			writeIndex++
		}
	}

	return arr[:writeIndex]
}

func RemoveRepeatedElementAndSort[T comparable](arr []T) []T {
	if len(arr) == 0 {
		return arr
	}

	sort.Slice(arr, func(i, j int) bool {
		return compareGeneric(arr[i], arr[j])
	})

	return RemoveRepeatedElementSorted(arr)
}

func compareGeneric[T comparable](a, b T) bool {

	switch any(a).(type) {
	case string:
		return any(a).(string) < any(b).(string)
	case int, int8, int16, int32, int64:
		return any(a).(int64) < any(b).(int64)
	case uint, uint8, uint16, uint32, uint64:
		return any(a).(uint64) < any(b).(uint64)
	case float32, float64:
		return any(a).(float64) < any(b).(float64)
	default:

		return false
	}
}

func RemoveRepeatedElementOptimized[T comparable](arr []T) []T {
	if len(arr) == 0 {
		return arr
	}

	if len(arr) <= 10 {
		return removeRepeatedSmallArray(arr)
	}

	return RemoveRepeatedElementGeneric(arr)
}

func removeRepeatedSmallArray[T comparable](arr []T) []T {
	result := make([]T, 0, len(arr))

	for i, item := range arr {
		found := false

		for j := 0; j < i; j++ {
			if arr[j] == item {
				found = true
				break
			}
		}
		if !found {
			result = append(result, item)
		}
	}

	return result
}

func RemoveRepeatedElementInPlace[T comparable](arr []T) int {
	if len(arr) <= 1 {
		return len(arr)
	}

	seen := make(map[T]bool, len(arr))
	writeIndex := 0

	for _, item := range arr {
		if !seen[item] {
			seen[item] = true
			arr[writeIndex] = item
			writeIndex++
		}
	}

	return writeIndex
}

func RemoveRepeatedElementWithCapacity[T comparable](arr []T, estimatedUniqueCount int) []T {
	if len(arr) == 0 {
		return arr
	}

	if estimatedUniqueCount <= 0 {
		estimatedUniqueCount = len(arr)
	}

	seen := make(map[T]bool, estimatedUniqueCount)
	result := make([]T, 0, estimatedUniqueCount)

	for _, item := range arr {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}

func StringSliceDedup(arr []string) []string {
	return RemoveRepeatedElementOptimized(arr)
}

func Uint64SliceDedup(arr []uint64) []uint64 {
	return RemoveRepeatedElementOptimized(arr)
}

func IntSliceDedup(arr []int) []int {
	return RemoveRepeatedElementOptimized(arr)
}

type DedupStats struct {
	OriginalCount   int
	UniqueCount     int
	DuplicateCount  int
	DuplicationRate float64
}

func RemoveRepeatedElementWithStats[T comparable](arr []T) ([]T, DedupStats) {
	originalCount := len(arr)
	result := RemoveRepeatedElementGeneric(arr)
	uniqueCount := len(result)
	duplicateCount := originalCount - uniqueCount

	var duplicationRate float64
	if originalCount > 0 {
		duplicationRate = float64(duplicateCount) / float64(originalCount) * 100
	}

	stats := DedupStats{
		OriginalCount:   originalCount,
		UniqueCount:     uniqueCount,
		DuplicateCount:  duplicateCount,
		DuplicationRate: duplicationRate,
	}

	return result, stats
}

func RemoveRepeatedElementBatch[T comparable](arrays ...[]T) [][]T {
	results := make([][]T, len(arrays))

	for i, arr := range arrays {
		results[i] = RemoveRepeatedElementGeneric(arr)
	}

	return results
}

func RemoveRepeatedElementParallel[T comparable](arr []T, numWorkers int) []T {
	if len(arr) == 0 || numWorkers <= 1 {
		return RemoveRepeatedElementGeneric(arr)
	}

	chunkSize := len(arr) / numWorkers
	if chunkSize == 0 {
		chunkSize = 1
	}

	type result struct {
		index int
		data  []T
	}

	resultChan := make(chan result, numWorkers)

	for i := 0; i < numWorkers; i++ {
		start := i * chunkSize
		end := start + chunkSize
		if i == numWorkers-1 {
			end = len(arr)
		}

		go func(index int, chunk []T) {
			dedupChunk := RemoveRepeatedElementGeneric(chunk)
			resultChan <- result{index: index, data: dedupChunk}
		}(i, arr[start:end])
	}

	chunks := make([][]T, numWorkers)
	for i := 0; i < numWorkers; i++ {
		res := <-resultChan
		chunks[res.index] = res.data
	}

	var merged []T
	for _, chunk := range chunks {
		merged = append(merged, chunk...)
	}

	return RemoveRepeatedElementGeneric(merged)
}
