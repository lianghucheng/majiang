package algorithm

import (
	"encoding/hex"
	"math/rand"
	"sort"
	"testing"
	"time"
)

func TestWin(t *testing.T) {
Label:
	for num := 0; num < 1; num++ {
		var card [][]int
		play := new(GDAnalyzer)
		play.RoomJokers = []int{11}
		rand.Seed(time.Now().UnixNano())
		result, _ := play.Win([]int{2, 21, 22, 23}, 1, false)
		t.Log(result)
		var a int
		for i := 0; i < 5; i++ {
			for k := 0; k < 1; k++ {
				b := make([]int, 0)
				for j := 0; j < 3*i+2; j++ {
					a = int(rand.Int31n(33))

					b = append(b, a)
				}
				sort.Ints(b)
				card = append(card, b)
			}
		}
		var result1, result2 bool
		for i := 0; i < len(card); i++ {
			length := len(card[i])
			data := length - 1
			result1, _ = play.Win(card[i][:data], card[i][length-1], false)
			count := 0
			for j := 0; j < len(card[i]); j++ {

				if card[i][j] == play.RoomJokers[0] {
					count++
					card[i][j] = 255
				}
			}
			sort.Ints(card[i])
			cards := IntTobyte(card[i])
			result2 = ExistHu3n2(cards)
			if result2 && !result1 {
				if count <= 3 {
					t.Log(card[i], cards)
					break Label
				}
			}
			if result1 && !result2 {
				if count <= 3 {
					t.Log(count, card[i], hex.EncodeToString(cards))
				}
			}
		}
	}
	/*
		for i := 0; i < len(card); i++ {


		}
	*/
}

func TestExistHu3n2(t *testing.T) {
	hu := ExistHu3n2([]byte{0x05, 0x06, 0x06, 0x07, 0x25, 0x27, 0xff, 0xff})
	t.Log(hu)
	play := new(GDAnalyzer)
	play.RoomJokers = []int{1}
	rand.Seed(time.Now().UnixNano())
	result, _ := play.Win([]int{5, 7, 23, 24, 31, 31, 1, 1, 33, 33, 21, 22, 23}, 1, false)
	t.Log(result)
}
func TestExistPingHu(t *testing.T) {
	hu := ExistPingHu([]byte{0x01, 0x02, 0x03, 0x4, 0x4})
	t.Log(hu)
}

/*
 2 3 1
 5 7 1
 23 24 1
 31 31
 33 33 1
*/

/*

3 3 11 12 14
16 16 17 24
26 255 255 255
 255

*/
