package main

import (
	"fmt"
	"io"
	"os"
)

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "错误:", err)
		os.Exit(1)
	}
}

func run(args []string, out io.Writer) error {
	if len(args) == 0 {
		printUsage(out)
		return fmt.Errorf("请给出颜色")
	}

	// 子命令：contrast 比对两个颜色，best 求最佳前景色，其余当转换。
	switch args[0] {
	case "contrast":
		if len(args) != 3 {
			return fmt.Errorf("contrast 需要两个颜色: go-color contrast <前景> <背景>")
		}
		fg, err := Parse(args[1])
		if err != nil {
			return err
		}
		bg, err := Parse(args[2])
		if err != nil {
			return err
		}
		ratio := Contrast(fg, bg)
		fmt.Fprintf(out, "对比度 %.2f:1\n", ratio)
		switch {
		case ratio >= 7:
			fmt.Fprintln(out, "AAA 级（最好，正文也够用）")
		case ratio >= 4.5:
			fmt.Fprintln(out, "AA 级（正文达标）")
		case ratio >= 3:
			fmt.Fprintln(out, "AA 级大字号（仅大标题够用）")
		default:
			fmt.Fprintln(out, "不达标，看不清")
		}
		return nil

	case "best":
		if len(args) != 2 {
			return fmt.Errorf("best 需要一个背景色: go-color best <背景>")
		}
		bg, err := Parse(args[1])
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "在 %s 上用 %s 最清楚（对比度 %.2f:1）\n",
			bg.Hex(), BestText(bg).Hex(), Contrast(bg, BestText(bg)))
		return nil
	}

	// 普通模式：把任意颜色转成三种表示都打出来。
	for _, a := range args {
		c, err := Parse(a)
		if err != nil {
			return err
		}
		h := c.ToHSL()
		fmt.Fprintf(out, "%s\n  HEX %s\n  RGB  %s\n  HSL  hsl(%.0f, %.0f%%, %.0f%%)\n",
			a, c.Hex(), c.String(), h.H, h.S, h.L)
	}
	return nil
}

func printUsage(out io.Writer) {
	fmt.Fprintln(out, `用法: go-color <颜色> [另一个颜色...]
      go-color contrast <前景> <背景>
      go-color best <背景>

颜色支持 #rgb / #rrggbb / rgb(r,g,b) / hsl(h,s%,l%) 形式。`)
}

