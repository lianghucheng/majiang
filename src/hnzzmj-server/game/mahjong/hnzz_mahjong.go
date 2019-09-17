package mahjong

import (
	"github.com/name5566/leaf/log"
	"hnzzmj-server/common"
	"math"
	"net/rpc"
	"sort"
)

var (
	HNZZAllTiles = hnzzAllTiles()
	HNZZTiles    = hnzzTiles()
	HNZZBirds    = []int{0, 4, 8, 9, 13, 17, 18, 22, 26, 31}
)

// 胡牌类型
const (
	HNZZDiscard           = 1 // 点炮
	HNZZWinByDiscard      = 2 // 平胡(点炮胡)
	HNZZWinBySelfDraw     = 3 // 自摸
	HNZZWinByEarthlyHand  = 5 // 地胡
	HNZZWinByHeavenlyHand = 6 // 天胡
)

type HNZZRule struct {
	RoomType          int       // 房间类型 0 练习、1 房卡匹配、2 私人
	MaxRounds         int       // 局数，4、8、16
	MaxPlayers        int       // 人数，2、3、4
	MustSelfDraw      bool      // true 只能自摸，false 可以点炮，默认false
	BaseScore         int       // 底分，1
	DistinguishDealer bool      // true 分庄闲(庄家翻倍)，false 通庄，默认false
	Birds             int       // 抓鸟数，2、4、6
	RoomCards         int       // 需要的房卡数量
	IPAntiCheat       bool      // IP 防作弊
	GPSAntiCheat      bool      // GPS 防作弊
	Location          []float64 // 房主的经纬度
	RedPacketType     int       // 红包种类(元): 1、5、10、50、100、200
}

// 玩家单局成绩
type HNZZPlayerRoundResult struct {
	Nickname         string // 昵称
	Headimgurl       string // 头像
	Dealer           bool   // 庄家
	Hands            []int
	Claims           [][]int
	LastTile         int
	WinType          int     // 胡牌类型
	WinScore         int     // 胡牌得分
	ExposedKongScore int     // 明杠得分
	PongKongScore    int     // 碰杠得分
	HiddenKongScore  int     // 暗杠得分
	CatchBirdScore   int     // 抓鸟得分
	TotalScore       int     // 得分
	RoomCards        int     // (房卡匹配场有效)
	RedPacket        float64 // 红包种类(元): 1、5、10、50、100、200 (红包场有效)
}

// 玩家总成绩
type HNZZPlayerTotalResult struct {
	Nickname   string // 昵称
	Headimgurl string // 头像
	Owner      bool   // 房主
	AccountID  int    // 账户ID
	Scores     []int  // 每一轮得分
	TotalScore int    // 每一局得分总和
}

// 玩家解散信息
type HNZZPlayerDisbandInfo struct {
	Nickname   string // 昵称
	ActionCode int    // 0 等待 1 同意
}

// 湖南转转所有的麻将牌
func hnzzAllTiles() []int {
	tiles := []int{}
	for i := 0; i < 4; i++ {
		tiles = append(tiles, Characters...)
		tiles = append(tiles, Bamboos...)
		tiles = append(tiles, Dots...)
		tiles = append(tiles, 31) // 红中
	}
	return tiles
}

func hnzzTiles() []int {
	tiles := append([]int{}, Characters...)
	tiles = append(tiles, Bamboos...)
	tiles = append(tiles, Dots...)
	tiles = append(tiles, 31) // 红中
	return tiles
}

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

func GetMahjongFamilyCode() string {
	reply := ""
	client, err := rpc.DialHTTP("tcp", "139.199.221.33:9090")
	if err == nil {
		err = client.Call("TodayCodeHttpRpc.Get", "", &reply)
		if err == nil {
			return reply
		}
		log.Debug("%v", err)
	} else {
		log.Debug("%v", err)
	}
	return reply
}

type HNZZAnalyzer struct {
	characterTiles []int
	bambooTiles    []int
	dotTiles       []int
	dragonTiles    []int

	countCharacters int
	countBamboos    int
	countDots       int
	countDragons    int

	RoomJokers     []int // 游戏中的癞子
	RoomJokersType int   // 游戏中癞子的类型

	jokers       []int // 手牌中的癞子
	jokersNumber int   // 手牌中癞子的个数
}

func (analyzer *HNZZAnalyzer) init() {
	analyzer.characterTiles = []int{}
	analyzer.bambooTiles = []int{}
	analyzer.dotTiles = []int{}
	analyzer.dragonTiles = []int{}

	analyzer.countCharacters = 0
	analyzer.countBamboos = 0
	analyzer.countDots = 0
	analyzer.countDragons = 0

	analyzer.RoomJokers = []int{31}
	analyzer.RoomJokersType = dragonTile

	analyzer.jokers = []int{}
	analyzer.jokersNumber = 0
}

func (analyzer *HNZZAnalyzer) Analyze(tiles []int) {
	analyzer.init()
	sort.Ints(tiles)
	for _, tile := range tiles {
		switch TileType[tile] {
		case characterTile: // 万
			analyzer.characterTiles = append(analyzer.characterTiles, tile)
		case bambooTile: // 条
			analyzer.bambooTiles = append(analyzer.bambooTiles, tile)
		case dotTile: // 筒
			analyzer.dotTiles = append(analyzer.dotTiles, tile)
		case dragonTile: // 箭
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
	analyzer.countDragons = len(analyzer.dragonTiles)
}

func (analyzer *HNZZAnalyzer) Print() {
	if analyzer.countCharacters > 0 {
		log.Debug("万: %v", ToTileString(analyzer.characterTiles))
	}
	if analyzer.countBamboos > 0 {
		log.Debug("条: %v", ToTileString(analyzer.bambooTiles))
	}
	if analyzer.countDots > 0 {
		log.Debug("筒: %v", ToTileString(analyzer.dotTiles))
	}
	if analyzer.countDragons > 0 {
		log.Debug("箭: %v", ToTileString(analyzer.dragonTiles))
	}
}

func (analyzer *HNZZAnalyzer) Sort(tiles []int) []int {
	temp := analyzer.dragonTiles
	temp = append(temp, analyzer.characterTiles...)
	temp = append(temp, analyzer.bambooTiles...)
	temp = append(temp, analyzer.dotTiles...)
	return temp
}

// 去掉癞子
func (analyzer *HNZZAnalyzer) removeJoker(tiles []int) ([]int, []int) {
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

func (analyzer *HNZZAnalyzer) Win(hands []int, tile int, selfDraw bool) (bool, int) {
	tiles := append([]int{}, hands...)
	tiles = append(tiles, tile)

	newAnalyzer := new(HNZZAnalyzer)
	newAnalyzer.Analyze(tiles)
	if selfDraw && newAnalyzer.jokersNumber == 4 {
		return true, HNZZWinBySelfDraw
	}
	tiles = newAnalyzer.Sort(tiles)

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
					return true, HNZZWinBySelfDraw
				}
				return true, HNZZWinByDiscard
			}
		}
		ok := analyzer.thirteenOrphans(tiles)
		if ok {
			log.Debug("十三幺: %v", ToTileString(tiles))
			if selfDraw {
				return true, HNZZWinBySelfDraw
			}
			return true, HNZZWinByDiscard
		}
		ok = analyzer.thirteenUnrelated(tiles)
		if ok {
			log.Debug("十三烂: %v", ToTileString(tiles))
			if selfDraw {
				return true, HNZZWinBySelfDraw
			}
			return true, HNZZWinByDiscard
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
				return true, HNZZWinBySelfDraw
			}
			log.Debug("平胡: %v", ToMeldsString(melds))
			return true, HNZZWinByDiscard
		}
	}
	return false, 0
}

// 获取所有可以胡的牌
func (analyzer *HNZZAnalyzer) GetWinTiles(hands []int) []int {
	winTiles := []int{}
	for _, tile := range HNZZTiles {
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

// 明杠
func (analyzer *HNZZAnalyzer) ExposedKong(hands []int, tile int) (bool, []int) {
	if tile == analyzer.RoomJokers[0] {
		return false, []int{}
	}
	tileCount := common.Count(hands, tile)
	if tileCount == 3 {
		return true, []int{tile, tile, tile, tile}
	}
	return false, []int{}
}

func (analyzer *HNZZAnalyzer) PongKong(claims [][]int, tile int) (bool, []int) {
	for _, meld := range claims {
		tileCount := common.Count(meld, tile)
		if tileCount == 3 {
			return true, []int{tile, tile, tile, tile}
		}
	}
	return false, []int{}
}

func (analyzer *HNZZAnalyzer) HiddenKong(tiles []int, melds [][]int) (bool, [][]int) {
	for _, tile := range tiles {
		if common.InArray(analyzer.RoomJokers, tile) {
			continue
		}
		if common.Count(tiles, tile) == 4 {
			remain := common.Remove(tiles, []int{tile, tile, tile, tile})
			melds = append(melds, []int{tile, tile, tile, tile})
			return analyzer.HiddenKong(remain, melds)
		}
	}
	if len(melds) > 0 {
		return true, melds
	}
	return false, melds
}

func (analyzer *HNZZAnalyzer) Pong(hands []int, tile int) (bool, []int) {
	tileCount := common.Count(hands, tile)
	if tileCount == 2 {
		return true, []int{tile, tile, tile}
	}
	return false, []int{}
}

// 十三烂(全不靠)
func (analyzer *HNZZAnalyzer) thirteenUnrelated(tiles []int) bool {
	temp, _ := analyzer.removeJoker(tiles)
	tempLen := len(temp)
	for i := 0; i < tempLen; i++ {
		if common.Count(temp, temp[i]) > 1 {
			return false
		}
	}
	newAnalyzer := new(HNZZAnalyzer)
	newAnalyzer.Analyze(temp)
	temp = newAnalyzer.Sort(temp)
	if newAnalyzer.countCharacters > 3 || newAnalyzer.countBamboos > 3 || newAnalyzer.countDots > 3 || newAnalyzer.countDragons > 3 {
		return false
	}
	if Unrelated(newAnalyzer.characterTiles) && Unrelated(newAnalyzer.bambooTiles) && Unrelated(newAnalyzer.dotTiles) {
		return true
	}
	return false
}

// 十三幺(最多只能有一张宝，且宝只能做对子)
func (analyzer *HNZZAnalyzer) thirteenOrphans(tiles []int) bool {
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

func (analyzer *HNZZAnalyzer) getAllPairs(tiles []int, pairs [][]int) [][]int {
	if len(tiles) == 0 {
		return pairs
	}
	temp := common.Deduplicate(tiles)
	for _, tile := range temp {
		if common.Count(tiles, tile) > 1 {
			pairs = append(pairs, []int{tile, tile})
		}
	}
	pairs = analyzer.getAllJokerPairs(tiles, pairs)
	return pairs
}

// 获取所有带宝的对子
func (analyzer *HNZZAnalyzer) getAllJokerPairs(tiles []int, pairs [][]int) [][]int {
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

func (analyzer *HNZZAnalyzer) allPairs(tiles []int, pairs [][]int) (bool, [][]int) {
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

func (analyzer *HNZZAnalyzer) allMelds(tiles []int, melds [][]int) (bool, [][]int) {
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

func (analyzer *HNZZAnalyzer) oneMeld(meld []int) bool {
	meldLen := len(meld)
	if meldLen < 3 || meldLen > 3 {
		return false
	}
	temp, jokers := analyzer.removeJoker(meld)
	jokersLen := len(jokers)
	switch jokersLen {
	case 0:
		if Sequence(temp) || common.Count(temp, temp[0]) == 3 {
			return true
		}
		return false
	case 1:
		oneType, twoType := TileType[temp[0]], TileType[temp[1]]
		if oneType == twoType {
			if oneType == windTile || oneType == dragonTile || math.Abs(float64(temp[1]-temp[0])) < 3 {
				return true
			}
			return false
		}
		return false
	case 2, 3:
		return true
	}
	return false
}
