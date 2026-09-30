package formation

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"testing"
	"time"
)

func TestAllUnitsPlaced(t *testing.T) {
	config := Formation{Columns: Columns, Spacing: Spacing}
	grid := NewGrid(nil)
	placed := config.Layout(Vec{}, 0, []Unit{{ID: "a"}, {ID: "b"}, {ID: "c"}}, grid)
	if len(placed) != 3 {
		t.Fatalf("三个单位应该有三个位置：%v", placed)
	}
}

func TestOriginTranslatesWholeLayout(t *testing.T) {
	config := Formation{Columns: Columns, Spacing: Spacing}
	grid := NewGrid(nil)
	units := []Unit{{ID: "a", Slot: 0}, {ID: "b", Slot: 1}, {ID: "c", Slot: 4}}
	atOrigin := config.Layout(Vec{}, 0, units, grid)
	moved := config.Layout(Vec{X: 10, Y: -4}, 0, units, grid)
	for _, unit := range units {
		delta := Vec{X: moved[unit.ID].X - atOrigin[unit.ID].X, Y: moved[unit.ID].Y - atOrigin[unit.ID].Y}
		if math.Abs(delta.X-10) > 1e-9 || math.Abs(delta.Y+4) > 1e-9 {
			t.Fatalf("整队平移后 %s 的位置偏移不对：%+v", unit.ID, delta)
		}
	}
}

func TestRotationKeepsUnitDistance(t *testing.T) {
	config := Formation{Columns: Columns, Spacing: Spacing}
	grid := NewGrid(nil)
	units := []Unit{{ID: "a", Slot: 0}, {ID: "b", Slot: 5}}
	flat := config.Layout(Vec{}, 0, units, grid)
	turned := config.Layout(Vec{}, 90, units, grid)
	distance := func(m map[string]Vec) float64 {
		return math.Hypot(m["a"].X-m["b"].X, m["a"].Y-m["b"].Y)
	}
	if math.Abs(distance(flat)-distance(turned)) > 1e-9 {
		t.Fatalf("旋转不该改变单位间距：%v %v", distance(flat), distance(turned))
	}
}

func TestConstants(t *testing.T) {
	if Columns != 3 || Spacing != 1.5 {
		t.Fatal("阵型常量被改了")
	}
}

func TestSpacingBetweenAdjacentSlots(t *testing.T) {
	config := Formation{Columns: Columns, Spacing: Spacing}
	grid := NewGrid(nil)
	units := []Unit{{ID: "u0", Slot: 0}, {ID: "u1", Slot: 1}, {ID: "u2", Slot: 2}}
	placed := config.Layout(Vec{}, 0, units, grid)
	for i, pair := range [][2]string{{"u0", "u1"}, {"u1", "u2"}} {
		distance := math.Hypot(placed[pair[0]].X-placed[pair[1]].X, placed[pair[0]].Y-placed[pair[1]].Y)
		if math.Abs(distance-Spacing) > 1e-9 {
			t.Fatalf("同排相邻 %d 距离应为 %v，实际 %v", i, Spacing, distance)
		}
	}
	if placed["u1"] != (Vec{X: 1.5, Y: 0}) || placed["u2"] != (Vec{X: 3, Y: 0}) {
		t.Fatalf("本地坐标必须乘 Spacing：%v", placed)
	}
}

func TestBlockedSlotShiftsWithinRowThenNextRow(t *testing.T) {
	config := Formation{Columns: Columns, Spacing: Spacing}
	grid := NewGrid([]Obstacle{{X: Spacing, Y: 0}})
	units := []Unit{{ID: "u0", Slot: 0}, {ID: "u1", Slot: 1}, {ID: "u2", Slot: 2}}
	placed := config.Layout(Vec{}, 0, units, grid)
	if placed["u1"] != (Vec{X: 3, Y: 0}) {
		t.Fatalf("被挡的单位应同排顺移到 3.00,0.00：%v", placed["u1"])
	}
	if placed["u2"] != (Vec{X: 0, Y: Spacing}) {
		t.Fatalf("排满后应换下一排第一列 0.00,1.50：%v", placed["u2"])
	}
	if grid.Checks() > 12 {
		t.Fatalf("checks 不应超过 12：%d", grid.Checks())
	}
	if grid.AtOriginCount() != 0 {
		t.Fatalf("不该有单位落不了位：%d", grid.AtOriginCount())
	}
}

func TestInputOrderDoesNotMatter(t *testing.T) {
	config := Formation{Columns: Columns, Spacing: Spacing}
	obstacles := []Obstacle{{X: Spacing, Y: 0}, {X: 0, Y: Spacing}}
	base := []Unit{{ID: "a", Slot: 0}, {ID: "b", Slot: 1}, {ID: "c", Slot: 2}, {ID: "d", Slot: 5}, {ID: "e", Slot: 7}}
	want := config.Layout(Vec{X: 2, Y: -1}, 30, base, NewGrid(obstacles))
	shuffled := []Unit{base[3], base[0], base[4], base[1], base[2]}
	got := config.Layout(Vec{X: 2, Y: -1}, 30, shuffled, NewGrid(obstacles))
	for _, unit := range base {
		if got[unit.ID] != want[unit.ID] {
			t.Fatalf("打乱输入顺序后 %s 位置变了：%v != %v", unit.ID, got[unit.ID], want[unit.ID])
		}
	}
}

func TestNoOverlapAndOriginFallbackObservable(t *testing.T) {
	config := Formation{Columns: Columns, Spacing: Spacing}
	// 槽位 0 那一排全被障碍物占住，下一排第一列也被占住。
	grid := NewGrid([]Obstacle{{X: 0, Y: 0}, {X: Spacing, Y: 0}, {X: 2 * Spacing, Y: 0}, {X: 0, Y: Spacing}})
	units := []Unit{{ID: "a", Slot: 0}, {ID: "b", Slot: 1}}
	placed := config.Layout(Vec{}, 0, units, grid)
	if placed["a"] != (Vec{}) || placed["b"] != (Vec{}) {
		t.Fatalf("找不到空位的单位应放在 origin：%v", placed)
	}
	if grid.AtOriginCount() != 2 {
		t.Fatalf("落不了位的单位数量应可观测：%d", grid.AtOriginCount())
	}
}

func TestAllPositionsDistinct(t *testing.T) {
	config := Formation{Columns: Columns, Spacing: Spacing}
	grid := NewGrid([]Obstacle{{X: Spacing, Y: 0}})
	units := []Unit{{ID: "a", Slot: 0}, {ID: "b", Slot: 1}, {ID: "c", Slot: 2}, {ID: "d", Slot: 3}, {ID: "e", Slot: 4}}
	placed := config.Layout(Vec{}, 0, units, grid)
	seen := make(map[Vec]string, len(placed))
	for id, position := range placed {
		if other, ok := seen[position]; ok {
			t.Fatalf("%s 与 %s 位置重叠：%v", id, other, position)
		}
		seen[position] = id
	}
}

// referenceLayout 是「按槽位算本地坐标 + 同排顺移避让」的独立参考实现，
// 用于随机对照：结果必须与 Layout 完全一致。
func referenceLayout(config Formation, origin Vec, headingDeg float64, units []Unit, obstacles []Obstacle) map[string]Vec {
	radians := headingDeg * math.Pi / 180
	cos, sin := math.Cos(radians), math.Sin(radians)
	blocked := func(position Vec) bool {
		for _, obstacle := range obstacles {
			if math.Abs(obstacle.X-position.X) < 1e-9 && math.Abs(obstacle.Y-position.Y) < 1e-9 {
				return true
			}
		}
		return false
	}
	ordered := append([]Unit(nil), units...)
	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].Slot != ordered[j].Slot {
			return ordered[i].Slot < ordered[j].Slot
		}
		return ordered[i].ID < ordered[j].ID
	})
	type cell struct{ row, col int }
	occupied := make(map[cell]bool, len(ordered))
	placed := make(map[string]Vec, len(ordered))
	for _, unit := range ordered {
		if unit.Slot < 0 {
			placed[unit.ID] = origin
			continue
		}
		row, col := unit.Slot/config.Columns, unit.Slot%config.Columns
		position := origin
		for candidate := col; candidate <= config.Columns; candidate++ {
			candidateRow, candidateCol := row, candidate
			if candidate == config.Columns {
				candidateRow, candidateCol = row+1, 0
			}
			localX := float64(candidateCol) * config.Spacing
			localY := float64(candidateRow) * config.Spacing
			candidatePosition := Vec{
				X: origin.X + localX*cos - localY*sin,
				Y: origin.Y + localX*sin + localY*cos,
			}
			if occupied[cell{candidateRow, candidateCol}] || blocked(candidatePosition) {
				continue
			}
			occupied[cell{candidateRow, candidateCol}] = true
			position = candidatePosition
			break
		}
		placed[unit.ID] = position
	}
	return placed
}

func TestLayoutMatchesReference(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	for trial := 0; trial < 200; trial++ {
		columns := 2 + rng.Intn(5)
		spacing := 0.5 + rng.Float64()*2.5
		heading := float64(rng.Intn(8)) * 45
		origin := Vec{X: rng.Float64()*20 - 10, Y: rng.Float64()*20 - 10}
		config := Formation{Columns: columns, Spacing: spacing}

		unitCount := 1 + rng.Intn(12)
		units := make([]Unit, unitCount)
		for i := range units {
			units[i] = Unit{ID: fmt.Sprintf("u%d", i), Slot: rng.Intn(4 * columns)}
		}
		// 打乱输入顺序，结果不应受影响。
		rng.Shuffle(len(units), func(i, j int) { units[i], units[j] = units[j], units[i] })

		// 障碍物一部分放在候选槽位点上（含旋转后的世界坐标），一部分是随机散点。
		obstacleCount := rng.Intn(2*unitCount + 1)
		obstacles := make([]Obstacle, 0, obstacleCount)
		radians := heading * math.Pi / 180
		cos, sin := math.Cos(radians), math.Sin(radians)
		for i := 0; i < obstacleCount; i++ {
			if rng.Intn(2) == 0 {
				slot := rng.Intn(5 * columns)
				localX := float64(slot%columns) * spacing
				localY := float64(slot/columns) * spacing
				obstacles = append(obstacles, Obstacle{
					X: origin.X + localX*cos - localY*sin,
					Y: origin.Y + localX*sin + localY*cos,
				})
			} else {
				obstacles = append(obstacles, Obstacle{
					X: origin.X + rng.Float64()*40 - 20,
					Y: origin.Y + rng.Float64()*40 - 20,
				})
			}
		}

		got := config.Layout(origin, heading, units, NewGrid(obstacles))
		want := referenceLayout(config, origin, heading, units, obstacles)
		if len(got) != len(want) {
			t.Fatalf("trial %d：单位数不一致 %d != %d", trial, len(got), len(want))
		}
		for id, wantPosition := range want {
			if got[id] != wantPosition {
				t.Fatalf("trial %d：%s 位置不一致 %v != %v", trial, id, got[id], wantPosition)
			}
		}
	}
}

func TestScaleStaysIndexed(t *testing.T) {
	config := Formation{Columns: 20, Spacing: Spacing}
	rng := rand.New(rand.NewSource(2))
	obstacles := make([]Obstacle, 200000)
	for i := range obstacles {
		obstacles[i] = Obstacle{X: rng.Float64() * 500, Y: rng.Float64() * 500}
	}
	units := make([]Unit, 2000)
	for i := range units {
		units[i] = Unit{ID: fmt.Sprintf("u%d", i), Slot: i}
	}
	rng.Shuffle(len(units), func(i, j int) { units[i], units[j] = units[j], units[i] })

	grid := NewGrid(obstacles)
	start := time.Now()
	placed := config.Layout(Vec{}, 0, units, grid)
	elapsed := time.Since(start)

	if len(placed) != len(units) {
		t.Fatalf("应落位 %d 个单位：%d", len(units), len(placed))
	}
	// 每单位最多探测 Columns+1 个候选格，每次探测一次障碍查询。
	if limit := len(units) * (config.Columns + 1); grid.Checks() > limit {
		t.Fatalf("障碍查询次数 %d 超过上限 %d", grid.Checks(), limit)
	}
	if elapsed > 5*time.Second {
		t.Fatalf("2000 单位 / 20 万障碍物布局耗时 %v，疑似退化成线性扫", elapsed)
	}
}
