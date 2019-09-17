package mahjong

const (
	WILDCARD = 0xff
	FENG     = 3
)

// 牌型满足 N * ABC + M *DDD +ＥＥ 形式
//cs 必须从小到大排序 鬼牌是牌里的最大值
//1.先移除对子，记下剩余牌的集合为Tn;
//2.针对每个Tn尝试移除一个顺子（ABC），成功转到2，失败到3。
//3.针对每个Tn尝试移除一个刻子（DDD），成功转到2。
//4.Tn为0，则表示，当前的方案可以胡牌；Tn不为0，则转到1移除下一个对子
func ExistHu3n2(cs []byte) bool {
	length := uint8(len(cs))
	if length == 0 {
		return false
	}
	var flag int32 //用来标记cs数组里的牌是否被移除
	for i := uint8(0); i < length; i++ {
		flag = flag | (1 << i)
	}
	var i, j uint8
	for i = 0; i < length-1; i++ {
		j = i + 1
		if cs[i] == cs[j] { // 两张非鬼牌或两张鬼牌
			f := removeEE(flag, i, j)
			if check3n(cs, f, length) {
				return true
			}
			if cs[i] == WILDCARD || j+1 < length && cs[j+1] != WILDCARD {
				i++
			}
		} else if cs[j] == WILDCARD { // 当c[i]不是鬼牌c[j]是鬼牌的时候
			for k := uint8(0); k < j; k++ {
				if k == 0 || cs[k] != cs[k-1] {
					f := removeEE(flag, j, k)
					if check3n(cs, f, length) {
						return true
					}
				}
			}
		}
	}
	return false
}

//移除将牌
func removeEE(flag int32, i uint8, j uint8) *int32 {
	flag = flag & (^(1 << i))
	flag = flag & (^(1 << j))
	return &flag
}

//检测是否满足(N * ABC + M *DDD)
func check3n(cs []byte, flag *int32, length uint8) bool {
	if *flag == 0 {
		return true
	}
	tempFlag := *flag
	if removeABC(cs, flag, length) {
		if check3n(cs, flag, length) {
			return true
		}
	}

	flag = &tempFlag
	if removeDDD(cs, flag, length) {
		if check3n(cs, flag, length) {
			return true
		}
	}
	return false
}

//尝试移除一个顺子
func removeABC(cs []byte, flag *int32, length uint8) bool {
	var first, second, third uint8
	firstFlag := false //是否存在合法的第一张牌
	for i := uint8(0); i < length; i++ {
		if (((*flag)>>i)&0x1) == 1 && (cs[i]>>4 < FENG || cs[i] == WILDCARD) {
			if !firstFlag {
				first = i
				firstFlag = true
			} else if second == 0 && (cs[i] == WILDCARD || cs[i] == cs[first]+1) {
				second = i
			} else if third == 0 && (cs[i] == WILDCARD || cs[i] == cs[first]+2) {
				third = i
			}
			if second > 0 && third > 0 {
				*flag = (*flag) & (^(1 << first))
				*flag = (*flag) & (^(1 << second))
				*flag = (*flag) & (^(1 << third))
				return true
			}
		}
	}
	return false
}

//判断一手风牌是否符合N * ABC + M *DDD
func removeFeng(cs []byte, flag *int32, length uint8) bool {
	singleNum := 0                   //单张数量(不成对数量)
	pairNum := 0                     //一对数量
	keZiNum := 0                     //刻子(3张相同)数量
	kongNum := 0                     //杠(4张相同)数量
	fengArrIndex := make([]uint8, 0) //风牌数组下标
	fengNum := 0                     //风牌数量
	for i := uint8(0); i < length; {
		if (((*flag)>>i)&0x1) == 1 && cs[i]>>4 == FENG {
			fengNum++
			fengArrIndex = append(fengArrIndex, i)
			sameCount := 0
			j := i + 1
			for ; j < length; j++ {
				if (((*flag)>>j)&0x1) == 1 && cs[j]>>4 == FENG && cs[j] == cs[i] {
					fengNum++
					fengArrIndex = append(fengArrIndex, j)
					sameCount++
				} else {
					break
				}
			}
			i = j
			switch sameCount {
			case 0:
				singleNum++
			case 1:
				pairNum++
			case 2:
				keZiNum++
			case 3:
				kongNum++
			}
		} else {
			i++
		}
	}
	if checkFeng3n(singleNum, pairNum, keZiNum, kongNum, fengNum) {
		for _, i := range fengArrIndex {
			*flag = (*flag) & (^(1 << i))
		}
		return true
	}
	return false
}

func checkFeng3n(singleNum, pairNum, keZiNum, kongNum, fengNum int) bool {
	if fengNum%3 != 0 { //数量不满足3n
		return false
	}
	if fengNum == 12 {
		return true
	}
	if fengNum == 3 {
		if singleNum == 3 || keZiNum == 1 {
			return true
		} else {
			return false
		}
	} else if fengNum == 6 {
		switch singleNum {
		case 0: //0单牌，3个对子或2个刻子才能成顺
			if pairNum == 3 || keZiNum == 2 {
				return true
			} else {
				return false
			}
		case 1: //仅有1张单牌 另外5张由1个对子和1个刻子组成，无法成顺
			return false
		case 2: //仅有2张单牌 另外四张要么是杠要么是2对，一定成顺
			return true
		case 3: //仅有3张单牌的话 另外3张一定是刻子
			return true
		}
	} else if fengNum == 9 {
		switch singleNum {
		case 0: //0单牌，牌型只有3种 111222333 112223333 112233444
			if kongNum == 1 { //112223333 这牌型不能做顺
				return false
			} else {
				return true
			}
		case 1: //仅有1张单牌 122333444(一个对子) 122334444(两个对子)  122224444(无对子)
			return true
		case 2: //仅有2张单牌 剩下7张由一个刻子加一个杠组成  123334444
			return true
		}
	} else if fengNum == 12 { //代码不会运行到这，前面已经提前返回true
		switch singleNum {
		case 0: //0单牌，牌型只有3种 111222333444(无对子) 112223334444(一个对子) 112233334444(两个对子)
			return true
		case 1: //仅有1张单牌 牌型只有1种 122233334444
			return true
		}
	}
	return false
}

//尝试移除一个刻子
func removeDDD(cs []byte, flag *int32, length uint8) bool {
	var count int8
	var first uint8 = 0xFF
	for i := uint8(0); i < length; i++ {
		if (((*flag) >> i) & 0x1) == 1 {
			if first == 0xFF {
				first = i
				*flag = (*flag) & (^(1 << i))
				continue
			}
			if cs[first] == cs[i] || cs[i] == WILDCARD {
				count++
				*flag = (*flag) & (^(1 << i))
			}
			if count == 2 {
				return true
			}
		}
	}
	return false
}

func check3nPengPeng(cs []byte, flag *int32, length uint8) bool {
	if *flag == 0 {
		return true
	}
	if removeDDD(cs, flag, length) {
		return check3nPengPeng(cs, flag, length)
	}
	return false
}
func checkABCPing(cs []byte, flag *int32, length uint8) bool {
	if *flag == 0 {
		return true
	}
	if removeABC(cs, flag, length) {
		return checkABCPing(cs, flag, length)
	}
	return false
}
func CardsUnique(cards []byte) []byte {
	var unique []byte
	for _, card := range cards {
		if !ByteInSlice(card, unique) {
			unique = append(unique, card)
		}
	}
	return unique
}

func Exist(c byte, cs []byte, n int) bool {
	for _, v := range cs {
		if n == 0 {
			return true
		}
		if c == v {
			n--
		}
	}
	return n == 0
}
func RemoveN(c byte, cs []byte, n int) []byte {
	for n > 0 {
		for i, v := range cs {
			if c == v {
				cs = append(cs[:i], cs[i+1:]...)
				break
			}
		}
		n--
	}
	return cs
}

func ByteInSlice(card byte, cards []byte) bool {
	for _, c := range cards {
		if c == card {
			return true
		}
	}
	return false
}

//是否存在某一组牌
func Contains(src []byte, sub []byte) bool {
	srcL := len(src)
	subL := len(sub)
	if srcL < subL {
		return false
	}
	n := 0
	flag := make([]bool, srcL)
	for _, c := range sub {
		for i := 0; i < srcL; i++ {
			if !flag[i] && c == src[i] {
				flag[i] = true
				n++
				if n == subL {
					return true
				}
				break
			}
		}
	}
	return n >= subL
}

// 对牌值从小到大排序，采用快速排序算法
func Sort(arr []byte, start, end int) {
	if start < end {
		i, j := start, end
		key := arr[(start+end)/2]
		for i <= j {
			for arr[i] < key {
				i++
			}
			for arr[j] > key {
				j--
			}
			if i <= j {
				arr[i], arr[j] = arr[j], arr[i]
				i++
				j--
			}
		}
		if start < j {
			Sort(arr, start, j)
		}
		if end > i {
			Sort(arr, i, end)
		}
	}
}

func IntTobyte(arr []int) []byte {
	bytes := make([]byte, 0)
	for i := 0; i < len(arr); i++ {
		if arr[i] >= 0 && arr[i] < 9 {
			bytes = append(bytes, byte(arr[i]+1))
		}
		if arr[i] >= 9 && arr[i] < 18 {
			bytes = append(bytes, byte(arr[i]+8))
		}
		if arr[i] >= 18 && arr[i] < 27 {
			bytes = append(bytes, byte(arr[i]+15))
		}
		if arr[i] >= 27 && arr[i] < 31 {
			bytes = append(bytes, byte(arr[i]+22))
		}
		if arr[i] >= 31 && arr[i] < 35 {
			bytes = append(bytes, byte(arr[i]+34))
		}
		if arr[i] == 255 {
			bytes = append(bytes, byte(0xff))
		}
	}
	return bytes
}
