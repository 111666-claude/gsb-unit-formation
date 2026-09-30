// Package formation 是阵型站位：按槽位算本地坐标、避让障碍物、同排顺移找空位。
package formation

import (
	"math"
	"sort"
)

// Columns 是每排列数。
const Columns = 3

// Spacing 是槽位间距，米。
const Spacing = 1.5

// cellSize 是障碍空间索引的格子边长，米。
const cellSize = 1.0

// Vec 是二维坐标。
type Vec struct {
	X, Y float64
}

// Unit 是一个单位，Slot 决定它站第几位。
type Unit struct {
	ID   string
	Slot int
}

// Obstacle 是一个障碍物。
type Obstacle struct {
	X, Y float64
}

// Grid 是障碍物索引。
type Grid struct {
	cells  map[[2]int64][]Obstacle
	checks int
}

// NewGrid 建索引：把障碍物按所在格子分桶，查询只看邻近格子。
func NewGrid(obstacles []Obstacle) *Grid {
	grid := &Grid{cells: make(map[[2]int64][]Obstacle, len(obstacles))}
	for _, obstacle := range obstacles {
		key := [2]int64{
			int64(math.Floor(obstacle.X / cellSize)),
			int64(math.Floor(obstacle.Y / cellSize)),
		}
		grid.cells[key] = append(grid.cells[key], obstacle)
	}
	return grid
}

// Blocked 判断一个位置是否被障碍物挡住。
// 只探测位置周围 3x3 的格子，代价与障碍物总数无关。
func (g *Grid) Blocked(position Vec) bool {
	g.checks++
	cx := int64(math.Floor(position.X / cellSize))
	cy := int64(math.Floor(position.Y / cellSize))
	for dx := int64(-1); dx <= 1; dx++ {
		for dy := int64(-1); dy <= 1; dy++ {
			for _, obstacle := range g.cells[[2]int64{cx + dx, cy + dy}] {
				if math.Abs(obstacle.X-position.X) < 1e-9 && math.Abs(obstacle.Y-position.Y) < 1e-9 {
					return true
				}
			}
		}
	}
	return false
}

// Checks 是障碍查询累计次数（规模观测）。
func (g *Grid) Checks() int { return g.checks }

// Formation 是阵型配置。
type Formation struct {
	Columns int
	Spacing float64
}

// Layout 算出每个单位的位置。
// 本地坐标由 Slot 决定：列是 Slot%Columns、行是 Slot/Columns，各乘 Spacing，
// 按 headingDeg 旋转后加上 origin。候选位置被障碍物或已占槽位挡住时，
// 在同一排顺移到下一个槽位，排尾换到下一排第一列（即槽位号加一）。
// 落位顺序按 Slot 升序，与输入顺序无关；找不到空位的单位放到 origin。
func (f Formation) Layout(origin Vec, headingDeg float64, units []Unit, grid *Grid) map[string]Vec {
	radians := headingDeg * math.Pi / 180
	cos, sin := math.Cos(radians), math.Sin(radians)
	position := func(slot int) Vec {
		localX := float64(slot%f.Columns) * f.Spacing
		localY := float64(slot/f.Columns) * f.Spacing
		return Vec{
			X: origin.X + localX*cos - localY*sin,
			Y: origin.Y + localX*sin + localY*cos,
		}
	}
	sorted := append([]Unit(nil), units...)
	sort.SliceStable(sorted, func(i, j int) bool { return sorted[i].Slot < sorted[j].Slot })
	occupied := make(map[int]struct{}, len(sorted))
	placed := make(map[string]Vec, len(sorted))
	for _, unit := range sorted {
		if f.Columns <= 0 {
			placed[unit.ID] = origin
			continue
		}
		slot := unit.Slot
		for {
			if _, taken := occupied[slot]; taken {
				slot++
				continue
			}
			candidate := position(slot)
			if grid != nil && grid.Blocked(candidate) {
				slot++
				continue
			}
			occupied[slot] = struct{}{}
			placed[unit.ID] = candidate
			break
		}
	}
	return placed
}
