package algorithm

import (
	"math"
	"sort"
	"util"

	"github.com/name5566/leaf/log"
)

var (
	GDAllTiles              = gdAllTiles()
	GDAllTilesWithoutHonors = gdAllTilesWithoutHonors()
	GDTiles                 = gdTiles()
	GDHorseTile             = []int{0, 4, 8, 9, 13, 17, 18, 22, 26, 27, 31}
)

// 胡牌类型
const (
	GDDiscard       = 1 // 点炮
	GDWinByDiscard  = 2 // 平胡(点炮胡)
	GDWinBySelfDraw = 3 // 自摸
)

// 广东所有的麻将牌
func gdAllTiles() []int {
	tiles := []int{}
	for i := 0; i < 4; i++ {
		tiles = append(tiles, Characters...) // 万
		tiles = append(tiles, Bamboos...)    // 条
		tiles = append(tiles, Dots...)       // 筒
		tiles = append(tiles, Winds...)      // 风
		tiles = append(tiles, Dragons...)    // 箭
	}
	return tiles
}

// 广东去掉字牌后的所有麻将牌
func gdAllTilesWithoutHonors() []int {
	tiles := []int{}
	for i := 0; i < 4; i++ {
		tiles = append(tiles, Characters...) // 万
		tiles = append(tiles, Bamboos...)    // 条
		tiles = append(tiles, Dots...)       // 筒
	}
	return tiles
}

func gdTiles() []int {
	tiles := append([]int{}, Characters...)
	tiles = append(tiles, Bamboos...)
	tiles = append(tiles, Dots...)
	tiles = append(tiles, Winds...)
	tiles = append(tiles, Dragons...)
	return tiles
}

func CatchHorse(x []int, y []int) (int, bool) {
	if len(x) < len(y) {
		return 0, false
	}

	count := 0
	for _, v1 := range y {
		for _, v2 := range x {
			if v2 == v1 {
				count++
			}
		}
	}
	if count == 0 {
		return 0, false
	}
	return count, true
}

func GetGDJokers(wildcard int) []int {
	wildCardType := TileType[wildcard]
	switch wildCardType {
	case characterTile: // 万
		jokerPos := (characterPositions[wildcard] + 1) % 9
		joker := Characters[jokerPos]
		return []int{joker}
	case bambooTile: // 条
		jokerPos := (bambooPositions[wildcard] + 1) % 9
		joker := Bamboos[jokerPos]
		return []int{joker}
	case dotTile: // 筒
		jokerPos := (dotPositions[wildcard] + 1) % 9
		joker := Dots[jokerPos]
		return []int{joker}
	case windTile: // 风
		jokerPos := (windPositions[wildcard] + 1) % 4
		joker := Winds[jokerPos]
		if windPositions[wildcard] == 3 {
			joker = Dragons[0]
		}
		return []int{joker}
	case dragonTile: // 箭
		jokerPos := (dragonPositions[wildcard] + 1) % 3
		joker := Dragons[jokerPos]
		if dragonPositions[wildcard] == 2 {
			joker = Winds[0]
		}
		return []int{joker}
	}
	log.Error("混儿: %v类型错误", wildcard)
	return []int{}
}

type GDAnalyzer struct {
	characterTiles []int //万牌
	bambooTiles    []int //条牌
	dotTiles       []int //饼牌
	windTiles      []int //风牌
	dragonTiles    []int //箭牌

	countCharacters int //风牌数量
	countBamboos    int //条牌数量
	countDots       int //饼牌数量
	countWinds      int //风牌数量
	countDragons    int //箭牌数量

	RoomJokers     []int // 游戏中的癞子
	RoomJokersType int   // 游戏中癞子的类型

	jokers       []int // 手牌中的癞子
	jokersNumber int   // 手牌中癞子的个数
}

func (analyzer *GDAnalyzer) init() {
	analyzer.characterTiles = []int{}
	analyzer.bambooTiles = []int{}
	analyzer.dotTiles = []int{}
	analyzer.windTiles = []int{}
	analyzer.dragonTiles = []int{}

	analyzer.countCharacters = 0
	analyzer.countBamboos = 0
	analyzer.countDots = 0
	analyzer.countWinds = 0
	analyzer.countDragons = 0

	analyzer.RoomJokers = []int{}
	analyzer.RoomJokersType = 0

	analyzer.jokers = []int{}
	analyzer.jokersNumber = 0
}

func (analyzer *GDAnalyzer) Analyze(tiles []int, roomJokers []int) {
	analyzer.init()
	analyzer.RoomJokers = roomJokers
	if len(roomJokers) > 0 {
		analyzer.RoomJokersType = TileType[roomJokers[0]]
	}
	sort.Ints(tiles)

	for _, tile := range tiles {
		switch TileType[tile] {
		case characterTile:
			analyzer.characterTiles = append(analyzer.characterTiles, tile)
		case bambooTile:
			analyzer.bambooTiles = append(analyzer.bambooTiles, tile)
		case dotTile:
			analyzer.dotTiles = append(analyzer.dotTiles, tile)
		case windTile:
			analyzer.windTiles = append(analyzer.windTiles, tile)
		case dragonTile:
			analyzer.dragonTiles = append(analyzer.dragonTiles, tile)
		}
		if util.Contain(analyzer.RoomJokers, []int{tile}) {
			analyzer.jokers = append(analyzer.jokers, tile)
			analyzer.jokersNumber++
		}
	}
	analyzer.countCharacters = len(analyzer.characterTiles)
	analyzer.countBamboos = len(analyzer.bambooTiles)
	analyzer.countDots = len(analyzer.dotTiles)
	analyzer.countWinds = len(analyzer.windTiles)
	analyzer.countDragons = len(analyzer.dragonTiles)
}

//排序
func (analyzer *GDAnalyzer) Sort() []int {
	temp := analyzer.characterTiles
	temp = append(temp, analyzer.bambooTiles...)
	temp = append(temp, analyzer.dotTiles...)
	temp = append(temp, analyzer.windTiles...)
	temp = append(temp, analyzer.dragonTiles...)
	return analyzer.reSort(temp)
}

func (analyzer *GDAnalyzer) reSort(tiles []int) []int {
	jokers := []int{}
	remain := []int{}
	for _, v := range tiles {
		if util.InArray(analyzer.RoomJokers, v) {
			jokers = append(jokers, v)
		} else {
			remain = append(remain, v)
		}
	}
	return append(jokers, remain...)
}

// 去掉癞子
func (analyzer *GDAnalyzer) removeJoker(tiles []int) ([]int, []int) {
	remain, jokers := []int{}, []int{}
	for _, tile := range tiles {
		if util.InArray(analyzer.RoomJokers, tile) {
			jokers = append(jokers, tile)
		} else {
			remain = append(remain, tile)
		}
	}
	return remain, jokers
}

func (analyzer *GDAnalyzer) Win(hands []int, tile int, selfDraw bool) (bool, int) {

	tiles := append([]int{}, hands...)
	tiles = append(tiles, tile)
	cards := append([]int{}, tiles...)

	newAnalyzer := new(GDAnalyzer)
	newAnalyzer.Analyze(tiles, analyzer.RoomJokers)
	if selfDraw && newAnalyzer.jokersNumber == 4 {
		return true, GDWinBySelfDraw
	}
	tiles = newAnalyzer.Sort()
	//把赖子转化成万能牌
	for i := 0; i < len(analyzer.RoomJokers); i++ {
		for j := 0; j < len(cards); j++ {
			if cards[j] == analyzer.RoomJokers[i] {
				cards[j] = 255
			}
		}
	}
	sort.Ints(cards)
	sliceCards := IntTobyte(cards)
	result1 := ExistQiDui(sliceCards)
	result2 := ExistHu3n2(sliceCards)
	if result1 || result2 {
		if selfDraw {
			return true, GDWinBySelfDraw
		}
		return true, GDWinByDiscard
	}
	return false, 0
}

/*
	pairs := analyzer.getAllPairs(tiles, [][]int{})
	if len(pairs) > 0 {
		// log.Debug("对子: %v", ToMeldsString(pairs))
	}
		tilesLen := len(tiles)
		if tilesLen == 14 {
			for _, pair := range pairs {
				temp := util.Remove(tiles, pair)
				ok, melds := analyzer.allPairs(temp, [][]int{})
				if ok { // 七小对
					melds = append(melds, pair)
					log.Debug("七小对: %v", ToMeldsString(melds))
					if selfDraw {
						return true, GDWinBySelfDraw
					}
					return true, GDWinByDiscard
				}
			}
		}
	for _, pair := range pairs {
		temp := util.Remove(tiles, pair)
		// log.Debug("%v, %v", ToTileString(pair), ToTileString(temp))
		ok, melds := analyzer.allMelds(temp, [][]int{})
		if ok {
			melds = append(melds, pair)
			if selfDraw {
				log.Debug("自摸: %v", ToMeldsString(melds))
				return true, GDWinBySelfDraw
			}
			log.Debug("平胡: %v", ToMeldsString(melds))
			return true, GDWinByDiscard
		}
	}
	return false, 0

}
*/
//获取所有对子
func (analyzer *GDAnalyzer) getAllPairs(tiles []int, pairs [][]int) [][]int {
	if len(tiles) == 0 {
		return pairs
	}
	temp := util.Deduplicate(tiles)
	for _, v := range temp {
		if util.Count(tiles, v) > 1 {
			pairs = append(pairs, []int{v, v})
		}
	}
	pairs = analyzer.getAllJokerPairs(tiles, pairs)
	return pairs
}

// 获取所有带宝的对子
func (analyzer *GDAnalyzer) getAllJokerPairs(tiles []int, pairs [][]int) [][]int {
	remain, jokers := analyzer.removeJoker(tiles)
	if len(remain) == 0 || len(jokers) == 0 {
		return pairs
	}
	remain = util.Deduplicate(remain)
	jokers = util.Deduplicate(jokers)
	remainLen, jokersLen := len(remain), len(jokers)
	for i := 0; i < jokersLen; i++ {
		for j := 0; j < remainLen; j++ {
			pairs = append(pairs, []int{jokers[i], remain[j]})
		}
	}
	return pairs
}

// 获取所有可以胡的牌
func (analyzer *GDAnalyzer) GetWinTiles(hands []int) []int {
	winTiles := []int{}
	for _, tile := range GDTiles {
		win, _ := analyzer.Win(hands, tile, true)
		if win {
			winTiles = append(winTiles, tile)
		}
	}
	if len(winTiles) > 0 {
		remain, jokers := analyzer.removeJoker(winTiles)
		return append(jokers, remain...)
	}
	return winTiles
}

//明杠
func (analyzer *GDAnalyzer) ExposedKong(hands []int, tile int) (bool, []int) {
	if len(analyzer.RoomJokers) > 0 {
		if tile == analyzer.RoomJokers[0] {
			return false, []int{}
		}
	}

	tileCount := util.Count(hands, tile)
	if tileCount == 3 {
		return true, []int{tile, tile, tile, tile}
	}
	return false, []int{}
}

//插杠
func (analyzer *GDAnalyzer) PongKong(claims [][]int, tile int) (bool, []int) {
	for _, meld := range claims {
		tileCount := util.Count(meld, tile)
		if tileCount == 3 {
			return true, []int{tile, tile, tile, tile}
		}
	}
	return false, []int{}
}

//暗杠
func (analyzer *GDAnalyzer) HiddenKong(tiles []int, melds [][]int) (bool, [][]int) {
	if len(tiles) < 4 {
		if len(melds) > 0 {
			return true, melds
		}
		return false, melds
	}
	for _, v := range tiles {
		if util.InArray(analyzer.RoomJokers, v) {
			continue
		}
		if util.Count(tiles, v) == 4 {
			remain := util.Remove(tiles, []int{v, v, v, v})
			melds = append(melds, []int{v, v, v, v})
			return analyzer.HiddenKong(remain, melds)
		}
	}
	if len(melds) > 0 {
		return true, melds
	}
	return false, melds
}

//碰
func (analyzer *GDAnalyzer) Pong(hands []int, tile int) (bool, []int) {
	tileCount := util.Count(hands, tile)
	if tileCount == 2 {
		return true, []int{tile, tile, tile}
	}
	return false, []int{}
}

func (analyzer *GDAnalyzer) allPairs(tiles []int, pairs [][]int) (bool, [][]int) {
	if len(tiles) == 0 {
		return true, pairs
	}
	remain, jokers := analyzer.removeJoker(tiles)
	jokersLen := len(jokers)
	if jokersLen == 2 {
		ok, pairs := analyzer.allPairs(remain, pairs)
		if ok {
			pairs = append(pairs, jokers)
			return true, pairs
		}
	}
	switch jokersLen {
	case 0:
		firstCount := util.Count(tiles, remain[0])
		switch firstCount {
		case 2, 4:
			pair := []int{remain[0], remain[0]}
			pairs = append(pairs, pair)
			temp := util.Remove(tiles, pair)
			return analyzer.allPairs(temp, pairs)
		}
		return false, pairs
	case 1, 2, 3:
		firstCount := util.Count(tiles, remain[0])
		switch firstCount {
		case 1, 3:
			pair := []int{jokers[0], remain[0]}
			pairs = append(pairs, pair)
			temp := util.Remove(tiles, pair)
			return analyzer.allPairs(temp, pairs)
		case 2, 4:
			pair := []int{remain[0], remain[0]}
			pairs = append(pairs, pair)
			temp := util.Remove(tiles, pair)
			return analyzer.allPairs(temp, pairs)
		}
		return false, pairs
	}
	return false, pairs
}

//所有刻子
func (analyzer *GDAnalyzer) allMelds(tiles []int, melds [][]int) (bool, [][]int) {
	if len(tiles) == 0 {
		return true, melds
	}
	remain, jokers := analyzer.removeJoker(tiles)
	jokersLen, remainLen := len(jokers), len(remain)
	if jokersLen == 2 {
		for i := 0; i < remainLen; i++ {
			meld := []int{jokers[0], jokers[1], remain[i]}
			temp := util.RemoveOnce(remain, remain[i])
			ok, newMelds := analyzer.allMelds(temp, [][]int{})
			if ok {
				newMelds = append(newMelds, meld)
				return true, newMelds
			}
		}
	} else if jokersLen == 3 {
		ok, newMelds := analyzer.allMelds(remain, [][]int{})
		if ok {
			newMelds = append(newMelds, jokers)
			return true, newMelds
		}
	}
	switch jokersLen {
	case 0:
		if remainLen > 3 {
			remain = util.Deduplicate(remain)
		} else if remainLen < 3 {
			return false, melds
		}
		firstCount := util.Count(tiles, remain[0])
		switch firstCount {
		case 1, 2, 4:
			if len(remain) < 3 {
				return false, melds
			}
			meld := remain[:3]
			if analyzer.oneMeld(meld) {
				melds = append(melds, meld)
				remain = util.Remove(tiles, meld)
				return analyzer.allMelds(remain, melds)
			}
		case 3:
			meld := []int{remain[0], remain[0], remain[0]}
			melds = append(melds, meld)
			remain = util.Remove(tiles, meld)
			return analyzer.allMelds(remain, melds)
		}
		return false, melds
	case 1, 2, 3:
		// log.Debug("%v", ToTileString(remain))
		for i := 0; i < remainLen; i++ {
			for j := i + 1; j < remainLen; j++ {
				meld := []int{jokers[0], remain[i], remain[j]}
				if analyzer.oneMeld(meld) {
					temp := util.Remove(tiles, meld)
					ok, newMelds := analyzer.allMelds(temp, [][]int{})
					if ok {
						newMelds = append(newMelds, meld)
						return true, newMelds
					}
				}
			}
		}
		return false, melds
	}
	return false, melds
}

//一个刻子
func (analyzer *GDAnalyzer) oneMeld(meld []int) bool {
	meldLen := len(meld)
	if meldLen < 3 || meldLen > 3 {
		return false
	}
	temp, jokers := analyzer.removeJoker(meld)
	jokersLen := len(jokers)
	switch jokersLen {
	case 0:
		// 27 代表东风
		if (meld[2] < 27 && Sequence(meld)) || util.Count(meld, meld[0]) == 3 {
			return true
		}
	case 1:
		oneType, twoType := TileType[temp[0]], TileType[temp[1]]
		if oneType != twoType {
			return false
		}
		if temp[1] < 27 { // 27 代表东风
			if math.Abs(float64(temp[1]-temp[0])) < 3 {
				return true
			}
		} else {
			if temp[0] == temp[1] {
				return true
			}
		}
	case 2, 3:
		return true
	}
	return false
}
