package mahjong

//判断是否风一色
//手头所有牌全风，不考虑牌型胡牌,即可不满足3n2
func ExistFengYiSe(cs []byte) bool {
	allFeng := true
	for _, v := range cs {
		if v != WILDCARD && v>>4 < FENG {
			allFeng = false
			break
		}
	}
	return allFeng
}
