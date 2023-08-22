package utils

import "sort"

// MaxInt64 获取多个数字中的最大数字
func MaxInt64(x int64, ys ...int64) int64 {
	ys = append(ys, x)
	sort.Slice(ys, func(i, j int) bool { return ys[i] > ys[j] })
	return ys[0]
}

// MinInt64 获取多个数字中的最小数字
func MinInt64(x int64, ys ...int64) int64 {
	ys = append(ys, x)
	sort.Slice(ys, func(i, j int) bool { return ys[i] < ys[j] })
	return ys[0]
}
