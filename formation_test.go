package formation

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"testing"
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

func TestRowNeighborSpacing(t *testing.T) {
	config := Formation{Columns: Columns, Spacing: Spacing}
	grid := NewGrid(nil)
	units := []Unit{{ID: "a", Slot: 0}, {ID: "b", Slot: 1}, {ID: "c", Slot: 2}}
	placed := config.Layout(Vec{}, 0, units, grid)
	for _, pair := range [][2]string{{"a", "b"}, {"b", "c"}} {
		distance := math.Hypot(placed[pair[0]].X-placed[pair[1]].X, placed[pair[0]].Y-placed[pair[1]].Y)
		if math.Abs(distance-Spacing) > 1e-9 {
			t.Fatalf("同排相邻单位距离必须等于 Spacing：%s-%s 是 %v", pair[0], pair[1], distance)
		}
	}
	if placed["b"] != (Vec{X: 1.5, Y: 0}) {
		t.Fatalf("槽位 1 应该在 1.50,0.00：%+v", placed["b"])
	}
}

func TestBlockedSlotShiftsAlongRow(t *testing.T) {
	config := Formation{Columns: Columns, Spacing: Spacing}
	grid := NewGrid([]Obstacle{{X: Spacing, Y: 0}})
	units := []Unit{{ID: "u0", Slot: 0}, {ID: "u1", Slot: 1}, {ID: "u2", Slot: 2}}
	placed := config.Layout(Vec{}, 0, units, grid)
	if placed["u1"] != (Vec{X: 3, Y: 0}) {
		t.Fatalf("被挡的单位应该同排顺移到 3.00,0.00：%+v", placed["u1"])
	}
	if placed["u2"] != (Vec{X: 0, Y: Spacing}) {
		t.Fatalf("被挤占的单位应该换到下一排第一列 0.00,1.50：%+v", placed["u2"])
	}
	if grid.Checks() > 12 {
		t.Fatalf("每单位只能探测常数次：checks=%d", grid.Checks())
	}
}

func TestInputOrderDoesNotMatter(t *testing.T) {
	config := Formation{Columns: Columns, Spacing: Spacing}
	obstacles := []Obstacle{{X: Spacing, Y: 0}, {X: 0, Y: Spacing}}
	forward := []Unit{{ID: "a", Slot: 0}, {ID: "b", Slot: 1}, {ID: "c", Slot: 2}, {ID: "d", Slot: 5}}
	shuffled := []Unit{{ID: "d", Slot: 5}, {ID: "b", Slot: 1}, {ID: "a", Slot: 0}, {ID: "c", Slot: 2}}
	first := config.Layout(Vec{X: 2, Y: 3}, 30, forward, NewGrid(obstacles))
	second := config.Layout(Vec{X: 2, Y: 3}, 30, shuffled, NewGrid(obstacles))
	for id, want := range first {
		if second[id] != want {
			t.Fatalf("列表顺序打乱后 %s 的位置变了：%+v != %+v", id, second[id], want)
		}
	}
}

func TestNoTwoUnitsSharePosition(t *testing.T) {
	config := Formation{Columns: Columns, Spacing: Spacing}
	grid := NewGrid([]Obstacle{{X: Spacing, Y: 0}, {X: 2 * Spacing, Y: 0}})
	units := []Unit{{ID: "a", Slot: 0}, {ID: "b", Slot: 1}, {ID: "c", Slot: 2}, {ID: "d", Slot: 3}}
	placed := config.Layout(Vec{}, 0, units, grid)
	seen := make(map[Vec]string, len(placed))
	for id, position := range placed {
		if other, dup := seen[position]; dup {
			t.Fatalf("%s 和 %s 站到了同一个位置 %+v", other, id, position)
		}
		seen[position] = id
	}
}

func TestUnplaceableUnitsGoToOrigin(t *testing.T) {
	config := Formation{Columns: 0, Spacing: Spacing}
	origin := Vec{X: 4, Y: -2}
	placed := config.Layout(origin, 0, []Unit{{ID: "a", Slot: 0}, {ID: "b", Slot: 1}}, NewGrid(nil))
	for id, position := range placed {
		if position != origin {
			t.Fatalf("找不到空位的 %s 应该放在 origin：%+v", id, position)
		}
	}
}

// referenceLayout 是测试里的参考实现：按槽位算本地坐标、线性扫障碍物、同排顺移避让。
func referenceLayout(config Formation, origin Vec, headingDeg float64, units []Unit, obstacles []Obstacle) map[string]Vec {
	radians := headingDeg * math.Pi / 180
	cos, sin := math.Cos(radians), math.Sin(radians)
	position := func(slot int) Vec {
		localX := float64(slot%config.Columns) * config.Spacing
		localY := float64(slot/config.Columns) * config.Spacing
		return Vec{
			X: origin.X + localX*cos - localY*sin,
			Y: origin.Y + localX*sin + localY*cos,
		}
	}
	blocked := func(candidate Vec) bool {
		for _, obstacle := range obstacles {
			if math.Abs(obstacle.X-candidate.X) < 1e-9 && math.Abs(obstacle.Y-candidate.Y) < 1e-9 {
				return true
			}
		}
		return false
	}
	sorted := append([]Unit(nil), units...)
	sort.SliceStable(sorted, func(i, j int) bool { return sorted[i].Slot < sorted[j].Slot })
	occupied := make(map[int]bool, len(sorted))
	placed := make(map[string]Vec, len(sorted))
	for _, unit := range sorted {
		slot := unit.Slot
		for occupied[slot] || blocked(position(slot)) {
			slot++
		}
		occupied[slot] = true
		placed[unit.ID] = position(slot)
	}
	return placed
}

func TestLayoutMatchesReference(t *testing.T) {
	rng := rand.New(rand.NewSource(20260930))
	for trial := 0; trial < 200; trial++ {
		config := Formation{Columns: 1 + rng.Intn(5), Spacing: 0.5 + rng.Float64()*2.5}
		origin := Vec{X: rng.Float64()*20 - 10, Y: rng.Float64()*20 - 10}
		heading := rng.Float64() * 360
		radians := heading * math.Pi / 180
		cos, sin := math.Cos(radians), math.Sin(radians)
		slotPosition := func(slot int) Vec {
			localX := float64(slot%config.Columns) * config.Spacing
			localY := float64(slot/config.Columns) * config.Spacing
			return Vec{
				X: origin.X + localX*cos - localY*sin,
				Y: origin.Y + localX*sin + localY*cos,
			}
		}
		unitCount := 1 + rng.Intn(30)
		slots := rng.Perm(4*unitCount + 8)[:unitCount]
		units := make([]Unit, unitCount)
		for index, slot := range slots {
			units[index] = Unit{ID: fmt.Sprintf("u%d", index), Slot: slot}
		}
		var obstacles []Obstacle
		for i := 0; i < rng.Intn(12); i++ {
			obstacles = append(obstacles, Obstacle(slotPosition(rng.Intn(4*unitCount+8))))
		}
		for i := 0; i < rng.Intn(20); i++ {
			obstacles = append(obstacles, Obstacle{
				X: origin.X + rng.Float64()*40 - 20,
				Y: origin.Y + rng.Float64()*40 - 20,
			})
		}
		got := config.Layout(origin, heading, units, NewGrid(obstacles))
		want := referenceLayout(config, origin, heading, units, obstacles)
		if len(got) != len(want) {
			t.Fatalf("第 %d 组：落位数量不一致 %d != %d", trial, len(got), len(want))
		}
		for id, wantPos := range want {
			if got[id] != wantPos {
				t.Fatalf("第 %d 组：%s 位置不一致 %+v != %+v", trial, id, got[id], wantPos)
			}
		}
	}
}

func TestLayoutScale(t *testing.T) {
	config := Formation{Columns: Columns, Spacing: Spacing}
	rng := rand.New(rand.NewSource(7))
	obstacles := make([]Obstacle, 200000)
	for i := range obstacles {
		obstacles[i] = Obstacle{X: rng.Float64() * 10000, Y: rng.Float64() * 10000}
	}
	grid := NewGrid(obstacles)
	units := make([]Unit, 2000)
	for i := range units {
		units[i] = Unit{ID: fmt.Sprintf("u%d", i), Slot: i}
	}
	placed := config.Layout(Vec{}, 0, units, grid)
	if len(placed) != len(units) {
		t.Fatalf("2000 个单位都要落位：%d", len(placed))
	}
	seen := make(map[Vec]struct{}, len(placed))
	for _, position := range placed {
		if _, dup := seen[position]; dup {
			t.Fatalf("大规模布局出现重叠位置 %+v", position)
		}
		seen[position] = struct{}{}
	}
	if limit := 4 * len(units); grid.Checks() > limit {
		t.Fatalf("每单位探测次数必须是常数：checks=%d 上限 %d", grid.Checks(), limit)
	}
}
