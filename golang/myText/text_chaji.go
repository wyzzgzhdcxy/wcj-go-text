package myText

import (
	"fmt"
	"github.com/wyzzgzhdcxy/wcj-go-common/core"
)

// Union 返回两个切片的并集
func Union(a, b []string) []string {
	fmt.Println(len(a), len(b))
	var results []string
	set := core.NewSet()

	for _, item := range a {
		set.Add(item)
		results = append(results, item)
	}

	for _, item := range b {
		if !set.Contains(item) {
			set.Add(item)
			results = append(results, item)
		}
	}
	return results
}

// Intersection 返回两个切片的交集
func Intersection(a, b []string) []string {
	fmt.Println(len(a), len(b))
	setB := make(map[string]bool)
	for _, item := range b {
		setB[item] = true
	}

	var result []string
	for _, item := range a {
		if setB[item] {
			result = append(result, item)
		}
	}
	return result
}

// Difference 返回a中存在但b中不存在的元素
func Difference(a, b []string) []string {
	fmt.Println(len(a), len(b))
	setB := make(map[string]bool)
	for _, item := range b {
		setB[item] = true
	}

	var result []string
	for _, item := range a {
		if !setB[item] {
			result = append(result, item)
		}
	}
	return result
}
