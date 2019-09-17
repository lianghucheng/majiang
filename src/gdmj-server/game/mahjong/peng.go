package mahjong

func ExistPingHu(cs []byte) bool {
	length := uint8(len(cs))
	var flag int32 //用来标记cs数组里的牌是否被移除
	for i := uint8(0); i < length; i++ {
		flag = flag | (1 << i)
	}
	var i, j uint8
	for i = 0; i < length-1; i++ {
		j = i + 1
		if cs[i] == cs[j] { // 两张非鬼牌或两张鬼牌
			f := removeEE(flag, i, j)
			if check3nPengPeng(cs, f, length) {
				return true
			}
			if cs[i] == WILDCARD || j+1 < length && cs[j+1] != WILDCARD {
				i++
			}
		} else if cs[j] == WILDCARD { // 当c[i]不是鬼牌c[j]是鬼牌的时候
			for k := uint8(0); k < j; k++ {
				if k == 0 || cs[k] != cs[k-1] {
					f := removeEE(flag, j, k)
					if check3nPengPeng(cs, f, length) {
						return true
					}
				}
			}
		}
	}
	return false
}
