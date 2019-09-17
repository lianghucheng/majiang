package mahjong

import (
	"gdmj-server/common"
)

const (
	_             = iota
	characterTile // 万
	bambooTile    // 条
	dotTile       // 筒
	windTile      // 风
	dragonTile    // 箭
	seasonTile    // 季
	flowerTile    // 花
)

// 动作码
const (
	ActionWin  = 0x01
	ActionKong = 0x02
	ActionPong = 0x04
	ActionChow = 0x08
)

// 杠牌类型
const (
	_           = iota
	ExposedKong // 明杠 点杠一家3番
	PongKong    // 碰杠 其他几家1番
	HiddenKong  // 暗杠 其他几家2番
)

// 游戏结果
const (
	ResultLose = iota // 0 失败
	ResultWin         // 1 胜利
	ResultDraw        // 2 流局
)

var (
	Characters = []int{0, 1, 2, 3, 4, 5, 6, 7, 8}          // 一到九万
	Bamboos    = []int{9, 10, 11, 12, 13, 14, 15, 16, 17}  // 一到九条
	Dots       = []int{18, 19, 20, 21, 22, 23, 24, 25, 26} // 一到九筒
	Winds      = []int{27, 28, 29, 30}                     // 东、南、西、北
	Dragons    = []int{31, 32, 33}                         // 中、发、白

	characterPositions = make(map[int]int) // 用于存放万的位置
	bambooPositions    = make(map[int]int) // 用于存放条的位置
	dotPositions       = make(map[int]int) // 用于存放筒的位置
	windPositions      = make(map[int]int) // 用于存放风的位置
	dragonPositions    = make(map[int]int) // 用于存放箭的位置

	TileType = []int{
		characterTile, characterTile, characterTile, characterTile, characterTile, characterTile, characterTile, characterTile, characterTile,
		bambooTile, bambooTile, bambooTile, bambooTile, bambooTile, bambooTile, bambooTile, bambooTile, bambooTile,
		dotTile, dotTile, dotTile, dotTile, dotTile, dotTile, dotTile, dotTile, dotTile,
		windTile, windTile, windTile, windTile,
		dragonTile, dragonTile, dragonTile,
	}
	TileString = []string{
		"一万", "二万", "三万", "四万", "五万", "六万", "七万", "八万", "九万",
		"一条", "二条", "三条", "四条", "五条", "六条", "七条", "八条", "九条",
		"一筒", "二筒", "三筒", "四筒", "五筒", "六筒", "七筒", "八筒", "九筒",
		"东风", "南风", "西风", "北风",
		"红中", "发财", "白板",
	}
)

func init() {
	for k, v := range Characters {
		characterPositions[v] = k
	}
	for k, v := range Bamboos {
		bambooPositions[v] = k
	}
	for k, v := range Dots {
		dotPositions[v] = k
	}
	for k, v := range Winds {
		windPositions[v] = k
	}
	for k, v := range Dragons {
		dragonPositions[v] = k
	}
}

func ToTileString(tiles []int) []string {
	s := []string{}
	for _, v := range tiles {
		s = append(s, TileString[v])
	}
	return s
}

func ToMeldsString(melds [][]int) [][]string {
	s := [][]string{}
	for _, v := range melds {
		s = append(s, ToTileString(v))
	}
	return s
}

func Sequence(tiles []int) bool {
	tilesLen := len(tiles)
	if tilesLen == 0 {
		return false
	}
	tile := tiles[0]
	for i := 1; i < tilesLen; i++ {
		tile2 := tiles[i]
		if TileType[tile2] == TileType[tile] && tile2-tile == 1 {
			tile = tile2
		} else {
			return false
		}
	}
	return true
}

func Quadruplet(tiles []int) bool {
	if len(tiles) == 4 {
		if common.Count(tiles, tiles[0]) == 4 {
			return true
		}
		return false
	}
	return false
}

// 去掉 melds 中带 tile 的刻子
func RemoveTriplet(melds [][]int, tile int) [][]int {
	newTiles := [][]int{}
	if TileType[tile] == seasonTile {
		for _, meld := range melds {
			if len(meld) == 3 && TileType[meld[0]] == seasonTile && TileType[meld[1]] == seasonTile && TileType[meld[2]] == seasonTile {
				continue
			}
			newTiles = append(newTiles, meld)
		}
		return newTiles
	}
	for _, meld := range melds {
		if common.Count(meld, tile) == 3 {
			continue
		}
		newTiles = append(newTiles, meld)
	}
	return newTiles
}
