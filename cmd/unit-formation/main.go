// Command unit-formation 跑阵型站位样例。
package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"example.com/formation"
)

func config() formation.Formation {
	return formation.Formation{Columns: formation.Columns, Spacing: formation.Spacing}
}

// Run 执行一次命令行调用，返回退出码。
func Run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("unit-formation", flag.ContinueOnError)
	flags.SetOutput(stderr)
	sample := flags.String("sample", "spacing", "spacing / shuffle / blocked")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	switch *sample {
	case "spacing":
		// 一排三个单位。
		grid := formation.NewGrid(nil)
		units := []formation.Unit{{ID: "u0", Slot: 0}, {ID: "u1", Slot: 1}, {ID: "u2", Slot: 2}}
		placed := config().Layout(formation.Vec{}, 0, units, grid)
		fmt.Fprintf(stdout, "u0=%.2f,%.2f u1=%.2f,%.2f u2=%.2f,%.2f\n",
			placed["u0"].X, placed["u0"].Y, placed["u1"].X, placed["u1"].Y, placed["u2"].X, placed["u2"].Y)
	case "shuffle":
		// 列表顺序与槽位顺序相反。
		grid := formation.NewGrid(nil)
		units := []formation.Unit{{ID: "u1", Slot: 1}, {ID: "u0", Slot: 0}}
		placed := config().Layout(formation.Vec{}, 0, units, grid)
		fmt.Fprintf(stdout, "u0=%.2f,%.2f u1=%.2f,%.2f\n",
			placed["u0"].X, placed["u0"].Y, placed["u1"].X, placed["u1"].Y)
	case "blocked":
		// 第二个槽位被障碍物占住。
		grid := formation.NewGrid([]formation.Obstacle{{X: formation.Spacing, Y: 0}})
		units := []formation.Unit{{ID: "u0", Slot: 0}, {ID: "u1", Slot: 1}, {ID: "u2", Slot: 2}}
		placed := config().Layout(formation.Vec{}, 0, units, grid)
		fmt.Fprintf(stdout, "u1=%.2f,%.2f u2=%.2f,%.2f checks=%d\n",
			placed["u1"].X, placed["u1"].Y, placed["u2"].X, placed["u2"].Y, grid.Checks())
	default:
		fmt.Fprintln(stderr, "需要 --sample spacing|shuffle|blocked")
		return 2
	}
	return 0
}

func main() {
	os.Exit(Run(os.Args[1:], os.Stdout, os.Stderr))
}
