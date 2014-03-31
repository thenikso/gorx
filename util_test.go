package rx

func processChan(testChan <-chan int, process func(int)) {
	for {
		select {
		case d := <-testChan:
			process(d)
		default:
			return
		}
	}
}

func countChan(testChan <-chan int) int {
	res := 0
	processChan(testChan, func(int) {
		res += 1
	})
	return res
}

func chanToSlice(testChan <-chan int) []int {
	res := []int{}
	processChan(testChan, func(i int) {
		res = append(res, i)
	})
	return res
}
