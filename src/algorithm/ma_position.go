package algorithm

func Position(dealposition, winposition, maxplay int) []int {
	/*
		Characters = []int{0, 1, 2, 3, 4, 5, 6, 7, 8}          // 一到九万
		Bamboos    = []int{9, 10, 11, 12, 13, 14, 15, 16, 17}  // 一到九条
		Dots       = []int{18, 19, 20, 21, 22, 23, 24, 25, 26} // 一到九筒
		Winds      = []int{27, 28, 29, 30}                     // 东、南、西、北
		Dragons    = []int{31, 32, 33}
	*/
	if maxplay == 2 {
		if dealposition == winposition {
			return []int{0, 2, 4, 6, 8, 9, 11, 13, 15, 17, 18, 20, 22, 24, 26, 27, 29, 31, 33}
		} else {
			return []int{1, 3, 5, 7, 10, 12, 14, 16, 19, 21, 23, 25, 28, 30, 32}
		}
	}

	if maxplay == 3 {
		if dealposition == winposition {
			return []int{0, 3, 6, 9, 12, 15, 27, 30, 31}
		}
		//如果是庄家得下家赢
		if (winposition+maxplay-dealposition)%maxplay == 1 {
			return []int{1, 4, 7, 10, 13, 16, 19, 22, 25, 28, 32}
		} else {
			return []int{2, 5, 8, 11, 14, 17, 20, 23, 26, 29, 33}
		}
	}
	if maxplay == 4 {
		index := (winposition + maxplay - dealposition) % maxplay
		switch index {
		case 0:
			return []int{0, 4, 8, 9, 13, 17, 18, 22, 26, 27, 31}
		case 1:
			return []int{1, 5, 10, 14, 19, 23, 28, 32}
		case 2:
			return []int{2, 6, 11, 15, 20, 24, 29, 33}
		case 3:
			return []int{3, 7, 12, 16, 21, 25, 30}
		}
	}
	return []int{}
}
