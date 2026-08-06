// go-color：颜色格式互转与对比度检查。
//
// 平时调 CSS、画图，经常要在 #rrggbb、rgb()、hsl() 之间来回倒腾，
// 还得确认白字配这个背景看不看得清。这个工具把这些活儿一次性做了，
// 支持 HEX / RGB / HSL 三种表示互转，以及 WCAG 对比度。
package main

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// RGB 是一个 0-255 的整数颜色。
type RGB struct {
	R, G, B uint8
}

// HSL 用 0-360 的色相和 0-100 的饱和度/亮度。
type HSL struct {
	H, S, L float64
}

// parseHex 解析 #rgb 或 #rrggbb，# 可有可无。
func parseHex(s string) (RGB, error) {
	t := strings.TrimSpace(s)
	t = strings.TrimPrefix(t, "#")
	switch len(t) {
	case 3:
		// #abc -> #aabbcc
		t = string([]byte{t[0], t[0], t[1], t[1], t[2], t[2]})
		fallthrough
	case 6:
		v, err := strconv.ParseUint(t, 16, 32)
		if err != nil {
			return RGB{}, fmt.Errorf("不是合法的十六进制颜色: %q", s)
		}
		return RGB{
			R: uint8(v >> 16),
			G: uint8(v >> 8),
			B: uint8(v),
		}, nil
	default:
		return RGB{}, fmt.Errorf("颜色长度不对，要 3 或 6 位十六进制: %q", s)
	}
}

// parseRGB 解析 "r,g,b" 或 "rgb(r, g, b)" 形式。
func parseRGB(s string) (RGB, error) {
	t := strings.TrimSpace(s)
	t = strings.TrimSuffix(strings.TrimPrefix(t, "rgb("), ")")
	parts := strings.Split(t, ",")
	if len(parts) != 3 {
		return RGB{}, fmt.Errorf("rgb 要三个分量: %q", s)
	}
	var c [3]uint8
	for i, p := range parts {
		n, err := strconv.Atoi(strings.TrimSpace(p))
		if err != nil {
			return RGB{}, fmt.Errorf("分量不是整数: %q", strings.TrimSpace(p))
		}
		if n < 0 || n > 255 {
			return RGB{}, fmt.Errorf("分量要 0-255，得到 %d", n)
		}
		c[i] = uint8(n)
	}
	return RGB{R: c[0], G: c[1], B: c[2]}, nil
}

// parseHSL 解析 "h,s,l" 或 "hsl(h, s%, l%)" 形式，s/l 带不带百分号都行。
func parseHSL(s string) (HSL, error) {
	t := strings.TrimSpace(s)
	t = strings.TrimSuffix(strings.TrimPrefix(t, "hsl("), ")")
	parts := strings.Split(t, ",")
	if len(parts) != 3 {
		return HSL{}, fmt.Errorf("hsl 要三个分量: %q", s)
	}
	trimPct := func(p string) string {
		return strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(p), "%"))
	}
	h, err := strconv.ParseFloat(trimPct(parts[0]), 64)
	if err != nil {
		return HSL{}, fmt.Errorf("色相不是整数: %q", parts[0])
	}
	ss, err := strconv.ParseFloat(trimPct(parts[1]), 64)
	if err != nil {
		return HSL{}, fmt.Errorf("饱和度不是整数: %q", parts[1])
	}
	l, err := strconv.ParseFloat(trimPct(parts[2]), 64)
	if err != nil {
		return HSL{}, fmt.Errorf("亮度不是整数: %q", parts[2])
	}
	if ss < 0 || ss > 100 || l < 0 || l > 100 {
		return HSL{}, fmt.Errorf("饱和度和亮度要 0-100")
	}
	// 色相超出 0-360 时卷绕回去，360 等同 0。
	h = math.Mod(h, 360)
	if h < 0 {
		h += 360
	}
	return HSL{H: h, S: ss, L: l}, nil
}

// ToHSL 把 RGB 转到 HSL。
func (c RGB) ToHSL() HSL {
	r := float64(c.R) / 255
	g := float64(c.G) / 255
	b := float64(c.B) / 255
	max := math.Max(r, math.Max(g, b))
	min := math.Min(r, math.Min(g, b))
	delta := max - min

	var h, s, l float64
	l = (max + min) / 2

	if delta == 0 {
		h, s = 0, 0
	} else {
		if l < 0.5 {
			s = delta / (max + min)
		} else {
			s = delta / (2 - max - min)
		}
		switch max {
		case r:
			h = (g - b) / delta
			if g < b {
				h += 6
			}
		case g:
			h = (b-r)/delta + 2
		default:
			h = (r-g)/delta + 4
		}
		h *= 60
	}
	return HSL{H: h, S: s * 100, L: l * 100}
}

// ToRGB 把 HSL 转回 RGB。
func (c HSL) ToRGB() RGB {
	h := c.H / 360
	s := c.S / 100
	l := c.L / 100
	if s == 0 {
		v := uint8(l * 255)
		return RGB{v, v, v}
	}
	var q float64
	if l < 0.5 {
		q = l * (1 + s)
	} else {
		q = l + s - l*s
	}
	p := 2*l - q
	hv := func(t float64) float64 {
		if t < 0 {
			t++
		}
		if t > 1 {
			t--
		}
		switch {
		case t < 1.0/6:
			return p + (q-p)*6*t
		case t < 0.5:
			return q
		case t < 2.0/3:
			return p + (q-p)*(2.0/3-t)*6
		default:
			return p
		}
	}
	return RGB{
		R: uint8(hv(h + 1.0/3) * 255),
		G: uint8(hv(h) * 255),
		B: uint8(hv(h - 1.0/3) * 255),
	}
}

// Hex 返回 #rrggbb 形式。
func (c RGB) Hex() string {
	return fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B)
}

// String 给 RGB 一个友好的展示。
func (c RGB) String() string {
	return fmt.Sprintf("rgb(%d, %d, %d)", c.R, c.G, c.B)
}

// relativeLuminance 按 WCAG 2.x 公式算相对亮度。
func (c RGB) relativeLuminance() float64 {
	f := func(v uint8) float64 {
		s := float64(v) / 255
		if s <= 0.03928 {
			return s / 12.92
		}
		return math.Pow((s+0.055)/1.055, 2.4)
	}
	r := f(c.R)
	g := f(c.G)
	b := f(c.B)
	return 0.2126*r + 0.7152*g + 0.0722*b
}

// Contrast 返回两个颜色的对比度（1.0 起，越大越清晰）。
func Contrast(a, b RGB) float64 {
	la := a.relativeLuminance()
	lb := b.relativeLuminance()
	hi := math.Max(la, lb)
	lo := math.Min(la, lb)
	return (hi + 0.05) / (lo + 0.05)
}

// BestText 返回在背景色上更清晰的前景（黑或白）。
func BestText(bg RGB) RGB {
	if Contrast(bg, RGB{255, 255, 255}) >= Contrast(bg, RGB{0, 0, 0}) {
		return RGB{255, 255, 255}
	}
	return RGB{0, 0, 0}
}

// Parse 尽力把任意支持的格式解析成 RGB。
// 顺序：hex -> rgb() -> hsl()，第一个成功的就返回。
func Parse(s string) (RGB, error) {
	if c, err := parseHex(s); err == nil {
		return c, nil
	}
	if strings.Contains(s, "rgb") {
		if c, err := parseRGB(s); err == nil {
			return c, nil
		}
	}
	if strings.Contains(s, "hsl") {
		if h, err := parseHSL(s); err == nil {
			return h.ToRGB(), nil
		}
	}
	// 再试一次 rgb 裸写法（不带 rgb 前缀）
	if c, err := parseRGB(s); err == nil {
		return c, nil
	}
	return RGB{}, fmt.Errorf("认不出这个颜色: %q", s)
}

