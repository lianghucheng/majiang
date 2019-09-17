package mahjong

import (
	"gdmj-server/common"
	"math"
	"sort"

	"github.com/name5566/leaf/log"
)

var (
	GDAllTiles              = gdAllTiles()
	GDAllTilesWithoutHonors = gdAllTilesWithoutHonors()
	GDTiles                 = gdTiles()
	GDHorseTile             = []int{0, 4, 8, 9, 13, 17, 18, 22, 26, 27, 31}
	HNZZBirds    = []int{0, 4, 8, 9, 13, 17, 18, 22, 26, 31}
)

// 胡牌类型
const (
	GDDiscard       = 1 // 点炮
	GDWinByDiscard  = 2 // 平胡(点炮胡)
	GDWinBySelfDraw = 3 // 自摸
)

//  胡牌类型---湖南转转麻将
const (
	HNZZDiscard           = 1 // 点炮
	HNZZWinByDiscard      = 2 // 平胡(点炮胡)
	HNZZWinBySelfDraw     = 3 // 自摸
	HNZZWinByEarthlyHand  = 5 // 地胡
	HNZZWinByHeavenlyHand = 6 // 天胡
)

//游戏类型
const (
	GD=1
	YA=2
	HNZZ=3
)


/*
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
*/
type GDRule struct {
	GameType       int
	Gun            bool      // 是否吓炮子
	RoomType       int       // 0 练习、1 房卡匹配场、2 私人房、 3 红包匹配、 4 红包私人
	MaxRounds      int       // 局数 4、8、16
	MaxPlayers     int       // 人数 2、3、4
	MustSelfDraw   bool      // true 只能自摸，false 可以点炮，默认false
	BaseScore      int       // 底分，1
	BuyHorse       int       // 买马 1匹马、2匹马
	WithHonors     bool      // 是否带风牌
	NeedJoker      bool      // 癞子
	RoomCards      int       // 需要的房卡数量
	IPAntiCheat    bool      // IP 防作弊
	GPSAntiCheat   bool      // GPS 防作弊
	RedPacketType  int       // 红包种类(元): 1、5、10、50、100、200
	Location       []float64 // 房主的经纬度
	RedDragonJoker bool      // 红中癞子

	DistinguishDealer bool      // true 分庄闲(庄家翻倍)，false 通庄，默认false---湖南转转麻将
	Birds             int       // 抓鸟数，2、4、6---湖南转转麻将
}

// 玩家单局成绩
type GDPlayerRoundResult struct {
	Nickname          string // 昵称
	Headimgurl        string // 头像
	Dealer            bool   // 庄家
	Hands             []int
	Claims            [][]int
	LastTile          int
	WinType           int     // 胡牌类型
	WinScore          int     // 胡牌得分
	CatchHorseScore   int     // 抓马得分
	ExposedKongScore  int     // 明杠得分
	PongKongScore     int     // 碰杠得分
	HiddenKongScore   int     // 暗杠得分
	TotalScore        int     // 总分
	RoomCards         int     // (房卡匹配场有效)
	RedPacket         float64 // 红包种类(元): 1、5、10、50、100、200 (红包场有效)
	GunScore          int     // 下炮子得分
	FollowDealerScore int     // 跟庄得分

	CatchBirdScore   int     // 抓鸟得分---湖南转转麻将
}

// 玩家总成绩
type GDPlayerTotalResult struct {
	Nickname   string // 昵称
	Headimgurl string // 头像
	Owner      bool   // 房主
	AccountID  int    // 账户ID
	Scores     []int  // 每一轮得分
	TotalScore int    // 每一局得分总和
}

// 玩家解散信息
type GDPlayerDisbandInfo struct {
	Nickname   string // 昵称
	ActionCode int    // 0 等待 1 同意
}

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
		if common.InArray(analyzer.RoomJokers, v) {
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
		if common.InArray(analyzer.RoomJokers, tile) {
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
				temp := common.Remove(tiles, pair)
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
		temp := common.Remove(tiles, pair)
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
func (analyzer *GDAnalyzer) getAllJokerPairs(tiles []int, pairs [][]int) [][]int {
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

	tileCount := common.Count(hands, tile)
	if tileCount == 3 {
		return true, []int{tile, tile, tile, tile}
	}
	return false, []int{}
}

//插杠
func (analyzer *GDAnalyzer) PongKong(claims [][]int, tile int) (bool, []int) {
	for _, meld := range claims {
		tileCount := common.Count(meld, tile)
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
func (analyzer *GDAnalyzer) Pong(hands []int, tile int) (bool, []int) {
	tileCount := common.Count(hands, tile)
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
func (analyzer *GDAnalyzer) allMelds(tiles []int, melds [][]int) (bool, [][]int) {
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



//湖南转转麻将
func CatchBird(tiles []int, n int) ([]int, []int) {
	tilesLen := len(tiles)
	if tilesLen == 0 {
		return tiles, tiles
	}
	if tilesLen < n {
		return tiles, common.GetSub(tiles, HNZZBirds)
	}
	temp := tiles[:n]
	return temp, common.GetSub(temp, HNZZBirds)
}