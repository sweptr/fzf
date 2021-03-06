package fzf

import (
	"math"
	"sort"
	"testing"

	"github.com/junegunn/fzf/src/curses"
	"github.com/junegunn/fzf/src/util"
)

func TestOffsetSort(t *testing.T) {
	offsets := []Offset{
		Offset{3, 5}, Offset{2, 7},
		Offset{1, 3}, Offset{2, 9}}
	sort.Sort(ByOrder(offsets))

	if offsets[0][0] != 1 || offsets[0][1] != 3 ||
		offsets[1][0] != 2 || offsets[1][1] != 7 ||
		offsets[2][0] != 2 || offsets[2][1] != 9 ||
		offsets[3][0] != 3 || offsets[3][1] != 5 {
		t.Error("Invalid order:", offsets)
	}
}

func TestRankComparison(t *testing.T) {
	rank := func(vals ...uint16) rank {
		return rank{
			points: [5]uint16{vals[0], 0, vals[1], vals[2], vals[3]},
			index:  int32(vals[4])}
	}
	if compareRanks(rank(3, 0, 0, 0, 5), rank(2, 0, 0, 0, 7), false) ||
		!compareRanks(rank(3, 0, 0, 0, 5), rank(3, 0, 0, 0, 6), false) ||
		!compareRanks(rank(1, 2, 0, 0, 3), rank(1, 3, 0, 0, 2), false) ||
		!compareRanks(rank(0, 0, 0, 0, 0), rank(0, 0, 0, 0, 0), false) {
		t.Error("Invalid order")
	}

	if compareRanks(rank(3, 0, 0, 0, 5), rank(2, 0, 0, 0, 7), true) ||
		!compareRanks(rank(3, 0, 0, 0, 5), rank(3, 0, 0, 0, 6), false) ||
		!compareRanks(rank(1, 2, 0, 0, 3), rank(1, 3, 0, 0, 2), true) ||
		!compareRanks(rank(0, 0, 0, 0, 0), rank(0, 0, 0, 0, 0), false) {
		t.Error("Invalid order (tac)")
	}
}

// Match length, string length, index
func TestResultRank(t *testing.T) {
	// FIXME global
	sortCriteria = []criterion{byMatchLen, byBonus, byLength}

	strs := [][]rune{[]rune("foo"), []rune("foobar"), []rune("bar"), []rune("baz")}
	item1 := buildResult(&Item{text: util.RunesToChars(strs[0]), index: 1}, []Offset{}, 2, 3)
	if item1.rank.points[0] != math.MaxUint16 || item1.rank.points[1] != math.MaxUint16-2 || item1.rank.points[2] != 3 || item1.item.index != 1 {
		t.Error(item1.rank)
	}
	// Only differ in index
	item2 := buildResult(&Item{text: util.RunesToChars(strs[0])}, []Offset{}, 2, 3)

	items := []*Result{item1, item2}
	sort.Sort(ByRelevance(items))
	if items[0] != item2 || items[1] != item1 {
		t.Error(items)
	}

	items = []*Result{item2, item1, item1, item2}
	sort.Sort(ByRelevance(items))
	if items[0] != item2 || items[1] != item2 ||
		items[2] != item1 || items[3] != item1 {
		t.Error(items, item1, item1.item.index, item2, item2.item.index)
	}

	// Sort by relevance
	item3 := buildResult(&Item{index: 2}, []Offset{Offset{1, 3}, Offset{5, 7}}, 0, 0)
	item4 := buildResult(&Item{index: 2}, []Offset{Offset{1, 2}, Offset{6, 7}}, 0, 0)
	item5 := buildResult(&Item{index: 2}, []Offset{Offset{1, 3}, Offset{5, 7}}, 0, 0)
	item6 := buildResult(&Item{index: 2}, []Offset{Offset{1, 2}, Offset{6, 7}}, 0, 0)
	items = []*Result{item1, item2, item3, item4, item5, item6}
	sort.Sort(ByRelevance(items))
	if items[0] != item6 || items[1] != item4 ||
		items[2] != item5 || items[3] != item3 ||
		items[4] != item2 || items[5] != item1 {
		t.Error(items)
	}
}

func TestColorOffset(t *testing.T) {
	// ------------ 20 ----  --  ----
	//   ++++++++        ++++++++++
	// --++++++++--    --++++++++++---

	offsets := []Offset{Offset{5, 15}, Offset{25, 35}}
	item := Result{
		item: &Item{
			colors: &[]ansiOffset{
				ansiOffset{[2]int32{0, 20}, ansiState{1, 5, false}},
				ansiOffset{[2]int32{22, 27}, ansiState{2, 6, true}},
				ansiOffset{[2]int32{30, 32}, ansiState{3, 7, false}},
				ansiOffset{[2]int32{33, 40}, ansiState{4, 8, true}}}}}
	// [{[0 5] 9 false} {[5 15] 99 false} {[15 20] 9 false} {[22 25] 10 true} {[25 35] 99 false} {[35 40] 11 true}]

	colors := item.colorOffsets(offsets, 99, false, true)
	assert := func(idx int, b int32, e int32, c int, bold bool) {
		o := colors[idx]
		if o.offset[0] != b || o.offset[1] != e || o.color != c || o.bold != bold {
			t.Error(o)
		}
	}
	assert(0, 0, 5, curses.ColUser, false)
	assert(1, 5, 15, 99, false)
	assert(2, 15, 20, curses.ColUser, false)
	assert(3, 22, 25, curses.ColUser+1, true)
	assert(4, 25, 35, 99, false)
	assert(5, 35, 40, curses.ColUser+2, true)
}
