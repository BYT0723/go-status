package util

var (
	sizeUnit = []string{"B", "K", "M", "G"}
)

// 计算大小，并返回处理后的值和单位
func CountSize(size uint64) (result float64, unit string) {
	result = float64(size)
	index := 0
	for result > 1024.0 {
		index++
		result /= 1024.0
	}
	unit = sizeUnit[index]
	return
}
