// Package formation 是阵型站位：把槽位换算成本地坐标，再按朝向旋转后加上队长位置。
// 缺陷：间距漏乘、槽位用的是输入下标，站位会重叠也会随输入顺序改变。
package formation

import "math"

// Columns 是默认的每排列数。
const Columns = 3

// Spacing 是默认的列间距，米。
const Spacing = 1.5

// Vec 是二维坐标。
type Vec struct {
	X, Y float64
}

// Unit 是一个单位。Slot 决定它站第几位。
type Unit struct {
	ID   string
	Slot int
}

// Formation 是阵型配置。
type Formation struct {
	Columns int
	Spacing float64
}

// Layout 算出每个单位的世界坐标。origin 是队长位置，headingDeg 是朝向（度，0 度朝 +X）。
// 缺陷：本地坐标没有乘 Spacing，行号列号用的是输入下标而不是 Unit.Slot。
func (f Formation) Layout(origin Vec, headingDeg float64, units []Unit) map[string]Vec {
	radians := headingDeg * math.Pi / 180
	cos, sin := math.Cos(radians), math.Sin(radians)
	placed := make(map[string]Vec, len(units))
	for index, unit := range units {
		localX := float64(index % f.Columns)
		localY := float64(index / f.Columns)
		placed[unit.ID] = Vec{
			X: origin.X + localX*cos - localY*sin,
			Y: origin.Y + localX*sin + localY*cos,
		}
	}
	return placed
}
