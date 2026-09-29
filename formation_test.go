package formation

import (
	"math"
	"testing"
)

func TestAllUnitsPlaced(t *testing.T) {
	config := Formation{Columns: Columns, Spacing: Spacing}
	placed := config.Layout(Vec{}, 0, []Unit{{ID: "a"}, {ID: "b"}, {ID: "c"}})
	if len(placed) != 3 {
		t.Fatalf("三个单位应该有三个位置：%v", placed)
	}
}

func TestOriginTranslatesWholeLayout(t *testing.T) {
	config := Formation{Columns: Columns, Spacing: Spacing}
	units := []Unit{{ID: "a", Slot: 0}, {ID: "b", Slot: 1}, {ID: "c", Slot: 4}}
	atOrigin := config.Layout(Vec{}, 0, units)
	moved := config.Layout(Vec{X: 10, Y: -4}, 0, units)
	for _, unit := range units {
		delta := Vec{X: moved[unit.ID].X - atOrigin[unit.ID].X, Y: moved[unit.ID].Y - atOrigin[unit.ID].Y}
		if math.Abs(delta.X-10) > 1e-9 || math.Abs(delta.Y+4) > 1e-9 {
			t.Fatalf("整队平移后 %s 的位置偏移不对：%+v", unit.ID, delta)
		}
	}
}

func TestRotationKeepsUnitDistance(t *testing.T) {
	config := Formation{Columns: Columns, Spacing: Spacing}
	units := []Unit{{ID: "a", Slot: 0}, {ID: "b", Slot: 5}}
	flat := config.Layout(Vec{}, 0, units)
	turned := config.Layout(Vec{}, 90, units)
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
