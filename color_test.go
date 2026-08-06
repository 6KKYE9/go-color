package main

import (
	"math"
	"testing"
)

func approx(a, b float64) bool { return math.Abs(a-b) < 1e-6 }

func TestParseHex(t *testing.T) {
	cases := []struct {
		in   string
		want RGB
	}{
		{"#ff0000", RGB{255, 0, 0}},
		{"00ff00", RGB{0, 255, 0}},
		{"#0f0", RGB{0, 255, 0}},
		{"#000000", RGB{0, 0, 0}},
		{"#ffffff", RGB{255, 255, 255}},
	}
	for _, c := range cases {
		got, err := parseHex(c.in)
		if err != nil {
			t.Errorf("parseHex(%q) 出错: %v", c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("parseHex(%q) = %v，想要 %v", c.in, got, c.want)
		}
	}
}

func TestParseHexBad(t *testing.T) {
	for _, bad := range []string{"#", "#12", "#gggggg", "red"} {
		if _, err := parseHex(bad); err == nil {
			t.Errorf("parseHex(%q) 应该报错", bad)
		}
	}
}

func TestParseRGB(t *testing.T) {
	cases := []struct {
		in   string
		want RGB
	}{
		{"rgb(10, 20, 30)", RGB{10, 20, 30}},
		{"10, 20, 30", RGB{10, 20, 30}},
		{"rgb(0,0,0)", RGB{0, 0, 0}},
		{"rgb(255, 255, 255)", RGB{255, 255, 255}},
	}
	for _, c := range cases {
		got, err := parseRGB(c.in)
		if err != nil {
			t.Errorf("parseRGB(%q) 出错: %v", c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("parseRGB(%q) = %v，想要 %v", c.in, got, c.want)
		}
	}
}

func TestParseRGBBad(t *testing.T) {
	for _, bad := range []string{"1, 2", "300, 0, 0", "a, b, c", "rgb(1,2)"} {
		if _, err := parseRGB(bad); err == nil {
			t.Errorf("parseRGB(%q) 应该报错", bad)
		}
	}
}

func TestParseHSL(t *testing.T) {
	cases := []struct {
		in   string
		want HSL
	}{
		{"hsl(0, 100%, 50%)", HSL{0, 100, 50}},
		{"hsl(120, 100, 50)", HSL{120, 100, 50}},
		{"0, 0, 0", HSL{0, 0, 0}},
		{"hsl(400, 50%, 50%)", HSL{40, 50, 50}}, // 400 卷绕到 40
		{"hsl(-20, 50%, 50%)", HSL{340, 50, 50}},
	}
	for _, c := range cases {
		got, err := parseHSL(c.in)
		if err != nil {
			t.Errorf("parseHSL(%q) 出错: %v", c.in, err)
			continue
		}
		if !approx(got.H, c.want.H) || !approx(got.S, c.want.S) || !approx(got.L, c.want.L) {
			t.Errorf("parseHSL(%q) = %+v，想要 %+v", c.in, got, c.want)
		}
	}
}

func TestParseHSLBad(t *testing.T) {
	for _, bad := range []string{"hsl(0, 200%, 50%)", "hsl(a, b, c)", "1, 2"} {
		if _, err := parseHSL(bad); err == nil {
			t.Errorf("parseHSL(%q) 应该报错", bad)
		}
	}
}

func TestRGBHSLRoundTrip(t *testing.T) {
	// 转过去再转回来，应该基本不丢精度。
	colors := []RGB{
		{255, 0, 0},
		{0, 128, 255},
		{123, 45, 67},
		{200, 200, 200},
	}
	for _, c := range colors {
		back := c.ToHSL().ToRGB()
		// 容忍 ±1 的取整误差
		if absDiff(c.R, back.R) > 1 || absDiff(c.G, back.G) > 1 || absDiff(c.B, back.B) > 1 {
			t.Errorf("%v 往返后变成 %v", c, back)
		}
	}
}

func TestContrastKnownValues(t *testing.T) {
	// 黑底白字对比度应为 21:1
	if r := Contrast(RGB{255, 255, 255}, RGB{0, 0, 0}); !approx(r, 21) {
		t.Errorf("黑白对比度 = %v，想要 21", r)
	}
	// 同色对比度为 1
	if r := Contrast(RGB{100, 100, 100}, RGB{100, 100, 100}); !approx(r, 1) {
		t.Errorf("同色对比度 = %v，想要 1", r)
	}
}

func TestBestText(t *testing.T) {
	if got := BestText(RGB{0, 0, 0}); got != (RGB{255, 255, 255}) {
		t.Errorf("黑底应该用白字，得到 %v", got)
	}
	if got := BestText(RGB{255, 255, 255}); got != (RGB{0, 0, 0}) {
		t.Errorf("白底应该用黑字，得到 %v", got)
	}
}

func TestParseDispatches(t *testing.T) {
	// 普通 Parse 应该按 hex -> rgb -> hsl 顺序认出
	if c, err := Parse("#abc"); err != nil || c != (RGB{170, 187, 204}) {
		t.Errorf("Parse(#abc) = %v, %v", c, err)
	}
	if c, err := Parse("rgb(1, 2, 3)"); err != nil || c != (RGB{1, 2, 3}) {
		t.Errorf("Parse(rgb) = %v, %v", c, err)
	}
	if c, err := Parse("hsl(0, 100%, 50%)"); err != nil || c != (RGB{255, 0, 0}) {
		t.Errorf("Parse(hsl) = %v, %v", c, err)
	}
	if _, err := Parse("完全认不出"); err == nil {
		t.Error("Parse 应该对乱码报错")
	}
}

func absDiff(a, b uint8) int {
	d := int(a) - int(b)
	if d < 0 {
		return -d
	}
	return d
}
