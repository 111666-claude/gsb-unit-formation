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

type cellKey struct{ x, y int64 }

// Grid 是障碍物的空间哈希索引。
type Grid struct {
	buckets  map[cellKey][]Obstacle
	checks   int
	atOrigin int
}

// NewGrid 建索引：障碍物按世界坐标的整数格分桶，内存 O(障碍物数)。
func NewGrid(obstacles []Obstacle) *Grid {
	g := &Grid{buckets: make(map[cellKey][]Obstacle, len(obstacles))}
	for _, obstacle := range obstacles {
		key := cellKey{int64(math.Floor(obstacle.X)), int64(math.Floor(obstacle.Y))}
		g.buckets[key] = append(g.buckets[key], obstacle)
	}
	return g
}

// Blocked 判断一个位置是否被障碍物挡住。
// 只探测落点周围 3x3 个整数格，探测代价与障碍物总数无关；每次调用计一次 Checks。
func (g *Grid) Blocked(position Vec) bool {
	g.checks++
	cx := int64(math.Floor(position.X))
	cy := int64(math.Floor(position.Y))
	for dx := int64(-1); dx <= 1; dx++ {
		for dy := int64(-1); dy <= 1; dy++ {
			for _, obstacle := range g.buckets[cellKey{cx + dx, cy + dy}] {
				if math.Abs(obstacle.X-position.X) < 1e-9 && math.Abs(obstacle.Y-position.Y) < 1e-9 {
					return true
				}
			}
		}
	}
	return false
}

// Checks 是最近一次布局的障碍查询累计次数（规模观测）。
func (g *Grid) Checks() int { return g.checks }

// AtOriginCount 是最近一次布局里找不到空位、被放到 origin 的单位数量。
func (g *Grid) AtOriginCount() int { return g.atOrigin }

// beginLayout 开始一次布局，重置本帧的查询计数与落不了位计数。
func (g *Grid) beginLayout() {
	g.checks = 0
	g.atOrigin = 0
}

// Formation 是阵型配置。
type Formation struct {
	Columns int
	Spacing float64
}

// Layout 算出每个单位的位置。
//
// 落位顺序按 Slot 升序（与输入顺序无关）；Slot 的本地坐标是
// (Slot%Columns, Slot/Columns) 乘 Spacing，再按朝向旋转、加 origin。
// 候选格被障碍物或已落位单位占住时，同排顺移到下一个槽位；
// 这一排走完后探一次下一排第一列，仍不可用就放到 origin。
func (f Formation) Layout(origin Vec, headingDeg float64, units []Unit, grid *Grid) map[string]Vec {
	grid.beginLayout()
	radians := headingDeg * math.Pi / 180
	cos, sin := math.Cos(radians), math.Sin(radians)

	world := func(row, col int) Vec {
		localX := float64(col) * f.Spacing
		localY := float64(row) * f.Spacing
		return Vec{
			X: origin.X + localX*cos - localY*sin,
			Y: origin.Y + localX*sin + localY*cos,
		}
	}

	ordered := append([]Unit(nil), units...)
	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].Slot != ordered[j].Slot {
			return ordered[i].Slot < ordered[j].Slot
		}
		return ordered[i].ID < ordered[j].ID
	})

	occupied := make(map[cellKey]bool, len(ordered))
	placed := make(map[string]Vec, len(ordered))
	for _, unit := range ordered {
		if unit.Slot < 0 {
			placed[unit.ID] = origin
			grid.atOrigin++
			continue
		}
		row, col := unit.Slot/f.Columns, unit.Slot%f.Columns

		// 同排内从自身槽位顺移到排尾，再探一次下一排第一列；仍不可用就放 origin。
		position := origin
		found := false
		for candidate := col; candidate <= f.Columns; candidate++ {
			candidateRow, candidateCol := row, candidate
			if candidate == f.Columns {
				candidateRow, candidateCol = row+1, 0
			}
			candidatePosition := world(candidateRow, candidateCol)
			key := cellKey{int64(candidateRow), int64(candidateCol)}
			if occupied[key] || grid.Blocked(candidatePosition) {
				continue
			}
			occupied[key] = true
			position = candidatePosition
			found = true
			break
		}
		if !found {
			grid.atOrigin++
		}
		placed[unit.ID] = position
	}
	return placed
}
