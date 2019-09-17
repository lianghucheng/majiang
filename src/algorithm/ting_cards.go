package algorithm

import (
	"gdmj-server/common"
)

func TingCards(handcards []int, remove int, roomJoker []int) []int {
	validcards := common.RemoveOnce(handcards, remove)

	result := make([]int, 0)
	for _, tile := range GDTiles {
		allcards := append(validcards, tile)

		for i := 0; i < len(roomJoker); i++ {
			for j := 0; j < len(allcards); j++ {
				if allcards[j] == roomJoker[i] {
					allcards[j] = 0xff
				}
			}
		}
		sliceCards := IntTobyte(allcards)
		Sort(sliceCards, 0, len(sliceCards)-1)
		if hu := ExistHu3n2(sliceCards); hu {
			result = append(result, tile)
		}
	}
	return result
}
