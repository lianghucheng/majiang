// 判断是否七对(7个对子)、龙七(1组4张)、双龙(2组4张)、三龙(3组4张)
// len(cs) == 14
// 有序
package mahjong

func ExistQiDui(cs []byte) bool {

	le := len(cs)
	if le != 14 {
		return false
	}
	singleCount := 0   //单张数量(不成对数量)
	pairCount := 0     //一对数量
	keZiCount := 0     //刻子(3张相同)数量
	kongCount := 0     //杠(4张相同)数量
	wildcardCount := 0 //鬼牌数量
	for i := 0; i < le; {
		if cs[i] == WILDCARD {
			wildcardCount++
			i++
		} else if i+1 < le && cs[i] == cs[i+1] {
			pairCount++
			if i+2 < le && cs[i] == cs[i+2] {
				if i+3 < le && cs[i] == cs[i+3] {
					kongCount++
				} else {
					keZiCount++
				}
			}
			i += 2
		} else {
			singleCount++
			i++
		}
	}
	if singleCount > wildcardCount {
		return false
	}
	kongCount += (wildcardCount - singleCount) / 2 //鬼牌补单牌后，剩余每对鬼牌补成杠
	kongCount += keZiCount                         //刻子补成杠
	if kongCount > 3 {
		kongCount = 3
	}

	if kongCount > 0 {
		return true
	}
	return true
}
