package mahjong

import (
	"testing"
)

func TestTingCards(t *testing.T) {
	card1 := []int{1, 2, 3, 5, 7}
	result1 := TingCards(card1, 7, []int{4})
	t.Log(result1)
}
