package mahjong

import (
	"github.com/name5566/leaf/log"
	"math"
	"sort"
	"yananmj-server/common"
)

var (
	YananAllTiles                     = yananAllTiles()
	YananAllTilesWithoutHonors        = yananAllTilesWithoutHonors()
	YananAllTilesWithoutHonorsWithRed = yananAllTilesWithoutHonorsWithRed()
	YananTiles                        = yananTiles()
)

//胡牌类型
const (
	_                  = iota
	YananWinByDiscard  // 1 平胡 //每家出1分
	YananWinBySelfDraw // 2 自摸 //每家出2分
	YananDiscard       // 3 点炮
)

// 延安麻将规则
type YananRule struct {
	RoomType       int       // 房间类型 0 练习、1 房卡匹配、2 私人
	MaxRounds      int       // 局数 8、16
	MaxPlayers     int       // 人数 4
	BaseScore      int       // 底分 1
	RoomCards      int       // 需要房卡数
	RedDragonJoker bool      // 红中癞子
	MustSelfDraw   bool      // true 只能自摸，false 可以点炮，默认false
	WithHonors     bool      // 是否带风牌
	Gun            bool      // 是否下炮子
	IPAntiCheat    bool      // IP 防作弊
	GPSAntiCheat   bool      // GPS 防作弊
	RedPacketType  int       // 红包种类(元): 1、5、10、50、100、200
	Location       []float64 // 房主的经纬度
}

// 玩家单局成绩
type YananPlayerRoundResult struct {
	Nickname          string // 昵称
	Headimgurl        string // 头像
	Dealer            bool   // 庄家
	Hands             []int
	Claims            [][]int
	LastTile          int
	Gun               int
	WinType           int     // 胡牌类型
	WinScore          int     // 胡牌得分
	ExposedKongScore  int     // 明杠得分
	PongKongScore     int     // 碰杠得分
	HiddenKongScore   int     // 暗杠得分
	GunScore          int     // 下炮子得分
	FollowDealerScore int     // 跟庄得分
	TotalScore        int     // 总分
	RoomCards         int     // (房卡匹配场有效)
	RedPacket         float64 // 红包种类(元): 1、5、10、50、100、200 (红包场有效)
}

// 玩家总成绩
type YananPlayerTotalResult struct {
	Nickname   string // 昵称
	Headimgurl string // 头像
	Owner      bool   // 房主
	AccountID  int    // 账户ID
	Scores     []int  // 每一轮得分
	TotalScore int    // 每一局得分总和
}

// 玩家解散者信息
type YananPlayerDisbandInfo struct {
	Nickname   string
	ActionCode int //0、等待 1、解散
}

//延安所有的麻将牌
func yananAllTiles() []int {
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

// 延安去掉字牌后的所有麻将牌
func yananAllTilesWithoutHonors() []int {
	tiles := []int{}
	for i := 0; i < 4; i++ {
		tiles = append(tiles, Characters...) // 万
		tiles = append(tiles, Bamboos...)    // 条
		tiles = append(tiles, Dots...)       // 筒
	}
	return tiles
}

// 延安去掉字牌带宝牌的所有麻将牌
func yananAllTilesWithoutHonorsWithRed() []int {
	tiles := []int{}
	for i := 0; i < 4; i++ {
		tiles = append(tiles, Characters...) // 万
		tiles = append(tiles, Bamboos...)    // 条
		tiles = append(tiles, Dots...)       // 筒
		tiles = append(tiles, 31)            // 红中
	}
	return tiles
}

func yananTiles() []int {
	tiles := append([]int{}, Characters...)
	tiles = append(tiles, Bamboos...)
	tiles = append(tiles, Dots...)
	tiles = append(tiles, Winds...)
	tiles = append(tiles, Dragons...)
	return tiles
}

type YananAnalyzer struct {
	characterTiles []int
	bambooTiles    []int
	dotTiles       []int
	windTiles      []int
	dragonTiles    []int

	countCharacters int
	countBamboos    int
	countDots       int
	countWinds      int
	countDragons    int

	RoomJokers     []int // 游戏中的癞子
	RoomJokersType int   // 游戏中癞子的类型

	jokers       []int // 手中的癞子
	jokersNumber int   // 手中癞子的个数
}

func (analyzer *YananAnalyzer) init() {
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

func (analyzer *YananAnalyzer) Analyze(tiles []int, roomJokers []int) {
	analyzer.init()
	analyzer.RoomJokers = roomJokers
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
		if common.Contain(analyzer.RoomJokers, []int{tile}) {
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
func (analyzer *YananAnalyzer) Sort() []int {
	temp := analyzer.characterTiles
	temp = append(temp, analyzer.bambooTiles...)
	temp = append(temp, analyzer.dotTiles...)
	temp = append(temp, analyzer.windTiles...)
	temp = append(temp, analyzer.dragonTiles...)
	return analyzer.reSort(temp)
}

func (analyzer *YananAnalyzer) reSort(tiles []int) []int {
	jokers := []int{}
	remain := []int{}
	for _, v := range tiles {
		if common.InArray(analyzer.RoomJokers, v) {
			jokers = append(jokers, v)
		} else {
			remain = append(remain, v)
		}
	}
	return append(jokers, remain...)
}

// 去掉宝牌
func (analyzer *YananAnalyzer) removeJoker(tiles []int) ([]int, []int) {
	remain, jokers := []int{}, []int{}
	for _, v := range tiles {
		if common.InArray(analyzer.RoomJokers, v) {
			jokers = append(jokers, v)
		} else {
			remain = append(remain, v)
		}
	}
	return remain, jokers
}

//胜利
func (analyzer *YananAnalyzer) Win(hands []int, tile int, selfDraw bool) (bool, int) {
	tiles := append([]int{}, hands...)
	tiles = append(tiles, tile)
	newAnalyzer := new(YananAnalyzer)
	newAnalyzer.Analyze(tiles, analyzer.RoomJokers)
	if selfDraw && newAnalyzer.jokersNumber == 4 {
		return true, YananWinBySelfDraw
	}
	tiles = newAnalyzer.Sort()

	pairs := analyzer.getAllPairs(tiles, [][]int{})
	tilesLen := len(tiles)
	if tilesLen == 14 {
		for _, pair := range pairs {
			temp := common.Remove(tiles, pair)
			ok, melds := analyzer.allPairs(temp, [][]int{})
			if ok { //七小对
				melds = append(melds, pair)
				log.Debug("七小对: %v", ToMeldsString(melds))
				if selfDraw {
					return true, YananWinBySelfDraw //自摸
				}
				return true, YananWinByDiscard //平胡
			}
		}

		ok := analyzer.thirteenOrphans(tiles)
		if ok {
			log.Debug("十三幺: %v", ToTileString(tiles))
			if selfDraw {
				return true, YananWinBySelfDraw
			}
			return true, YananWinByDiscard
		}
		ok = analyzer.thirteenUnrelated(tiles)
		if ok {
			log.Debug("十三烂: %v", ToTileString(tiles))
			if selfDraw {
				return true, YananWinBySelfDraw
			}
			return true, YananWinByDiscard
		}
	}

	for _, pair := range pairs {
		temp := common.Remove(tiles, pair)
		ok, melds := analyzer.allMelds(temp, [][]int{})
		if ok {
			melds = append(melds, pair)
			if selfDraw {
				log.Debug("自摸: %v", ToMeldsString(melds))
				return true, YananWinBySelfDraw
			}
			log.Debug("平胡: %v", ToMeldsString(melds))
			return true, YananWinByDiscard
		}
	}
	return false, 0
}

//获取所有可以胡的牌
func (analyzer *YananAnalyzer) GetWinTiles(hands []int) []int {
	winTiles := []int{}
	for _, tile := range YananTiles {
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
func (analyzer *YananAnalyzer) ExposedKong(hands []int, tile int) (bool, []int) {
	if len(analyzer.RoomJokers) > 0 {
		if tile == analyzer.RoomJokers[0] {
			return false, []int{}
		}
	}

	tileCount := common.Count(hands, tile)
	if tileCount == 3 {
		return true, []int{tile, tile, tile, tile}
	}
	return false, []int{}
}

//碰杠
func (analyzer *YananAnalyzer) PongKong(claims [][]int, tile int) (bool, []int) {
	for _, meld := range claims {
		tileCount := common.Count(meld, tile)
		if tileCount == 3 {
			return true, []int{tile, tile, tile, tile}
		}
	}
	return false, []int{}
}

//暗杠
func (analyzer *YananAnalyzer) HiddenKong(tiles []int, melds [][]int) (bool, [][]int) {
	for _, v := range tiles {
		if common.InArray(analyzer.RoomJokers, v) {
			continue
		}
		if common.Count(tiles, v) == 4 {
			remain := common.Remove(tiles, []int{v, v, v, v})
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
func (analyzer *YananAnalyzer) Pong(hands []int, tile int) (bool, []int) {
	tileCount := common.Count(hands, tile)
	if tileCount == 2 {
		return true, []int{tile, tile, tile}
	}
	return false, []int{}
}

// 十三烂(全不靠)
func (analyzer *YananAnalyzer) thirteenUnrelated(tiles []int) bool {
	temp, _ := analyzer.removeJoker(tiles)
	tempLen := len(temp)
	for i := 0; i < tempLen; i++ {
		if common.Count(temp, temp[i]) > 1 {
			return false
		}
	}
	newAnalyzer := new(YananAnalyzer)
	newAnalyzer.Analyze(temp, analyzer.RoomJokers)
	temp = newAnalyzer.Sort()
	if newAnalyzer.countCharacters > 3 || newAnalyzer.countBamboos > 3 || newAnalyzer.countDots > 3 || newAnalyzer.countDragons > 3 {
		return false
	}
	if Unrelated(newAnalyzer.characterTiles) && Unrelated(newAnalyzer.bambooTiles) && Unrelated(newAnalyzer.dotTiles) {
		return true
	}
	return false
}

// 十三幺(最多只能有一张宝，且宝只能做对子)
func (analyzer *YananAnalyzer) thirteenOrphans(tiles []int) bool {
	remain, _ := analyzer.removeJoker(tiles)
	remainLen := len(remain)
	switch remainLen {
	case 13, 14:
		remain = common.Deduplicate(remain)
		if common.Equal(remain, TerminalAndHonour) {
			return true
		}
		return false
	}
	return false
}

//获取所有对子
func (analyzer *YananAnalyzer) getAllPairs(tiles []int, pairs [][]int) [][]int {
	if len(tiles) == 0 {
		return pairs
	}
	temp := common.Deduplicate(tiles)
	for _, v := range temp {
		if common.Count(tiles, v) > 1 {
			pairs = append(pairs, []int{v, v})
		}
	}
	pairs = analyzer.getAllJokerPairs(tiles, pairs)
	return pairs
}

// 获取所有带宝的对子
func (analyzer *YananAnalyzer) getAllJokerPairs(tiles []int, pairs [][]int) [][]int {
	remain, jokers := analyzer.removeJoker(tiles)
	if len(remain) == 0 || len(jokers) == 0 {
		return pairs
	}
	remain = common.Deduplicate(remain)
	jokers = common.Deduplicate(jokers)
	remainLen, jokersLen := len(remain), len(jokers)
	for i := 0; i < jokersLen; i++ {
		for j := 0; j < remainLen; j++ {
			pairs = append(pairs, []int{jokers[i], remain[j]})
		}
	}
	return pairs
}

//所有的对子
func (analyzer *YananAnalyzer) allPairs(tiles []int, pairs [][]int) (bool, [][]int) {
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
		firstCount := common.Count(tiles, remain[0])
		switch firstCount {
		case 2, 4:
			pair := []int{remain[0], remain[0]}
			pairs = append(pairs, pair)
			temp := common.Remove(tiles, pair)
			return analyzer.allPairs(temp, pairs)
		}
		return false, pairs
	case 1, 2, 3:
		firstCount := common.Count(tiles, remain[0])
		switch firstCount {
		case 1, 3:
			pair := []int{jokers[0], remain[0]}
			pairs = append(pairs, pair)
			temp := common.Remove(tiles, pair)
			return analyzer.allPairs(temp, pairs)
		case 2, 4:
			pair := []int{remain[0], remain[0]}
			pairs = append(pairs, pair)
			temp := common.Remove(tiles, pair)
			return analyzer.allPairs(temp, pairs)
		}
		return false, pairs
	}
	return false, pairs
}

//所有刻子
func (analyzer *YananAnalyzer) allMelds(tiles []int, melds [][]int) (bool, [][]int) {
	if len(tiles) == 0 {
		return true, melds
	}
	remain, jokers := analyzer.removeJoker(tiles)
	jokersLen, remainLen := len(jokers), len(remain)
	if jokersLen == 2 {
		for i := 0; i < remainLen; i++ {
			meld := []int{jokers[0], jokers[1], remain[i]}
			temp := common.RemoveOnce(remain, remain[i])
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
			remain = common.Deduplicate(remain)
		} else if remainLen < 3 {
			return false, melds
		}
		firstCount := common.Count(tiles, remain[0])
		switch firstCount {
		case 1, 2, 4:
			if len(remain) < 3 {
				return false, melds
			}
			meld := remain[:3]
			if analyzer.oneMeld(meld) {
				melds = append(melds, meld)
				remain = common.Remove(tiles, meld)
				return analyzer.allMelds(remain, melds)
			}
		case 3:
			meld := []int{remain[0], remain[0], remain[0]}
			melds = append(melds, meld)
			remain = common.Remove(tiles, meld)
			return analyzer.allMelds(remain, melds)
		}
		return false, melds
	case 1, 2, 3:
		// log.Debug("%v", ToTileString(remain))
		for i := 0; i < remainLen; i++ {
			for j := i + 1; j < remainLen; j++ {
				meld := []int{jokers[0], remain[i], remain[j]}
				if analyzer.oneMeld(meld) {
					temp := common.Remove(tiles, meld)
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

//一个组合
func (analyzer *YananAnalyzer) oneMeld(meld []int) bool {
	meldLen := len(meld)
	if meldLen != 3 {
		return false
	}
	temp, jokers := analyzer.removeJoker(meld)
	jokersLen := len(jokers)
	switch jokersLen {
	case 0:
		// 27 代表东风
		if (meld[2] < 27 && Sequence(meld)) || common.Count(meld, meld[0]) == 3 {
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
