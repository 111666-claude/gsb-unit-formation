// Command unit-formation 跑阵型站位样例。
package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"example.com/formation"
)

// Run 执行一次命令行调用，返回退出码。
func Run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("unit-formation", flag.ContinueOnError)
	flags.SetOutput(stderr)
	sample := flags.String("sample", "spacing", "spacing / shuffle")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	config := formation.Formation{Columns: formation.Columns, Spacing: formation.Spacing}
	var units []formation.Unit
	switch *sample {
	case "spacing":
		// 一排三个单位。
		units = []formation.Unit{{ID: "u0", Slot: 0}, {ID: "u1", Slot: 1}, {ID: "u2", Slot: 2}}
	case "shuffle":
		// 同一排两个单位，列表顺序与槽位顺序相反。
		units = []formation.Unit{{ID: "u1", Slot: 1}, {ID: "u0", Slot: 0}}
	default:
		fmt.Fprintln(stderr, "需要 --sample spacing|shuffle")
		return 2
	}
	placed := config.Layout(formation.Vec{}, 0, units)
	switch *sample {
	case "spacing":
		fmt.Fprintf(stdout, "u0.x=%.2f u1.x=%.2f u2.x=%.2f\n",
			placed["u0"].X, placed["u1"].X, placed["u2"].X)
	case "shuffle":
		fmt.Fprintf(stdout, "u0.x=%.2f u1.x=%.2f\n", placed["u0"].X, placed["u1"].X)
	}
	return 0
}

func main() {
	os.Exit(Run(os.Args[1:], os.Stdout, os.Stderr))
}
