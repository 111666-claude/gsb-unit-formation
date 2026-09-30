// Package formation 是阵型站位：按槽位算本地坐标、避让障碍物、同排顺移找空位。
// 缺陷：本地坐标漏乘 Spacing、行列用输入下标而不是 Slot、障碍物与槽位冲突都没处理、
// 障碍查询是线性扫。
package formation

import "math"

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

// Grid 是障碍物索引。
type Grid struct {
	obstacles []Obstacle
	checks    int
}

// NewGrid 建索引。
// 缺陷：只把障碍物存下来，没有建任何空间索引。
func NewGrid(obstacles []Obstacle) *Grid {
	return &Grid{obstacles: append([]Obstacle(nil), obstacles...)}
}

// Blocked 判断一个位置是否被障碍物挡住。
// 缺陷：逐个障碍物线性比较，代价随障碍物数量增长。
func (g *Grid) Blocked(position Vec) bool {
	for _, obstacle := range g.obstacles {
		g.checks++
		if math.Abs(obstacle.X-position.X) < 1e-9 && math.Abs(obstacle.Y-position.Y) < 1e-9 {
			return true
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
// 缺陷：本地坐标没有乘 Spacing、行列号用的是输入下标而不是 Unit.Slot，
// 障碍物与槽位冲突都被完全忽略。
func (f Formation) Layout(origin Vec, headingDeg float64, units []Unit, grid *Grid) map[string]Vec {
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
