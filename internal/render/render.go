// Package render 使用 OpenType 字体与矢量圆角卡片绘制三级帮助菜单 PNG。
//
// [ParseFonts] 接收调用方提供的字体字节；[Index]、[Detail]、[Command] 接收已本地化的文案。
// 每次绘制独占 font.Face，解析后的 [Fonts] 可供多个请求共用。
package render

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
	"golang.org/x/image/vector"
)

const (
	imageWidth       = 1080
	margin           = 48
	gap              = 24
	columns          = 3
	padding          = 28
	badgeWidth       = 52
	bodySize         = 23
	boldSize         = 28
	titleSize        = 42
	lineHeight       = 34
	boldLineHeight   = 40
	titleLineHeight  = 58
	cardRadius       = 20
	badgeRadius      = 14
	accentWidth      = 6
	accentInset      = 6
	accentRadius     = 3
	fontDPI          = 72
	descriptionLines = 2
)

// Fonts 保存解析后的常规与粗体字体。
type Fonts struct{ Regular, Bold *opentype.Font }

// ParseFonts 解析调用方提供的 OTF 或 TTF 字节。
func ParseFonts(regular, bold []byte) (*Fonts, error) {
	r, err := opentype.Parse(regular)
	if err != nil {
		return nil, fmt.Errorf("parse regular font: %w", err)
	}
	b, err := opentype.Parse(bold)
	if err != nil {
		return nil, fmt.Errorf("parse bold font: %w", err)
	}
	return &Fonts{Regular: r, Bold: b}, nil
}

// LoadFonts 从目录读取随包分发的 Noto Sans SC 字体。
func LoadFonts(dir string) (*Fonts, error) {
	regular, err := os.ReadFile(filepath.Join(dir, "NotoSansSC-Regular.otf"))
	if err != nil {
		return nil, fmt.Errorf("read regular font: %w", err)
	}
	bold, err := os.ReadFile(filepath.Join(dir, "NotoSansSC-Bold.otf"))
	if err != nil {
		return nil, fmt.Errorf("read bold font: %w", err)
	}
	return ParseFonts(regular, bold)
}

// Theme 指定背景、卡片、边框与文字配色。
type Theme struct{ Background, Card, Border, Text, Muted, Accent, Badge color.NRGBA }

var (
	// Light 提供浅灰背景、白色卡片与蓝色强调。
	Light = Theme{Background: hex(0xf2f5fa), Card: hex(0xffffff), Border: hex(0xdce3ed), Text: hex(0x172b4d), Muted: hex(0x62748e), Accent: hex(0x316bea), Badge: hex(0xeaf0ff)}
	// Dark 提供深蓝灰背景与高对比文字。
	Dark = Theme{Background: hex(0x111927), Card: hex(0x1c283b), Border: hex(0x34435b), Text: hex(0xeff4ff), Muted: hex(0xa6b7d0), Accent: hex(0x8cb4ff), Badge: hex(0x293f63)}
)

func hex(value uint32) color.NRGBA {
	return color.NRGBA{R: uint8(value >> 16), G: uint8(value >> 8), B: uint8(value), A: 255}
}

// PluginCard 提供一级卡片的完整名称、说明和本地化命令数量。
type PluginCard struct{ Name, Description, Count string }

// IndexData 提供一级总览文案与已排序的插件。
type IndexData struct {
	Title, Subtitle string
	Plugins         []PluginCard
}

// CommandRow 提供二级命令的用法、别名与说明。
type CommandRow struct{ Usage, Aliases, Description string }

// DetailData 提供插件头部与其全部命令。
type DetailData struct {
	Title, Description, Subtitle string
	Commands                     []CommandRow
}

// CommandData 提供三级详情的字段标题与完整内容。
type CommandData struct {
	Title, Usage, Aliases, Description, Plugin              string
	UsageLabel, AliasesLabel, DescriptionLabel, PluginLabel string
}

type canvas struct {
	theme             Theme
	body, bold, title font.Face
	image             *image.RGBA
}

func newCanvas(f *Fonts, theme Theme) (*canvas, error) {
	c := &canvas{theme: theme}
	for _, face := range []struct {
		target *font.Face
		source *opentype.Font
		size   float64
	}{
		{&c.body, f.Regular, bodySize}, {&c.bold, f.Bold, boldSize}, {&c.title, f.Bold, titleSize},
	} {
		loaded, err := opentype.NewFace(face.source, &opentype.FaceOptions{Size: face.size, DPI: fontDPI, Hinting: font.HintingNone})
		if err != nil {
			c.close()
			return nil, fmt.Errorf("create menu font face: %w", err)
		}
		*face.target = loaded
	}
	return c, nil
}

func (c *canvas) close() {
	for _, face := range []font.Face{c.body, c.bold, c.title} {
		if face != nil {
			// opentype.Face.Close 固定返回 nil。
			_ = face.Close()
		}
	}
}

func (c *canvas) start(height int) {
	c.image = image.NewRGBA(image.Rect(0, 0, imageWidth, height))
	draw.Draw(c.image, c.image.Bounds(), image.NewUniform(c.theme.Background), image.Point{}, draw.Src)
}

func (c *canvas) encode() ([]byte, error) {
	var out bytes.Buffer
	if err := png.Encode(&out, c.image); err != nil {
		return nil, fmt.Errorf("encode menu PNG: %w", err)
	}
	return out.Bytes(), nil
}

// wrap 按字符测量，保留显式换行，避免中英文长词越过卡片边缘。
func wrap(face font.Face, text string, width int) []string {
	var lines []string
	for _, paragraph := range strings.Split(text, "\n") {
		line := ""
		for _, r := range paragraph {
			next := line + string(r)
			if line != "" && font.MeasureString(face, next).Ceil() > width {
				lines = append(lines, line)
				line = string(r)
			} else {
				line = next
			}
		}
		lines = append(lines, line)
	}
	return lines
}

func ellipsis(face font.Face, lines []string, width, limit int) []string {
	if len(lines) <= limit {
		return lines
	}
	lines = lines[:limit]
	last := []rune(lines[limit-1])
	for len(last) > 0 && font.MeasureString(face, string(last)+"…").Ceil() > width {
		last = last[:len(last)-1]
	}
	lines[limit-1] = string(last) + "…"
	return lines
}

func (c *canvas) text(face font.Face, lines []string, x, y, height int, ink color.NRGBA) int {
	d := font.Drawer{Dst: c.image, Src: image.NewUniform(ink), Face: face}
	for _, line := range lines {
		d.Dot = fixed.P(x, y+face.Metrics().Ascent.Ceil())
		d.DrawString(line)
		y += height
	}
	return y
}

func (c *canvas) rounded(x, y, width, height, radius int, ink color.NRGBA) {
	// 三次贝塞尔近似四分之一圆，vector 在卡片边缘执行抗锯齿。
	const circleControl = 0.55228475
	x0, y0, x1, y1, r := float32(x), float32(y), float32(x+width), float32(y+height), float32(radius)
	k := r * circleControl
	v := vector.NewRasterizer(imageWidth, c.image.Bounds().Dy())
	v.MoveTo(x0+r, y0)
	v.LineTo(x1-r, y0)
	v.CubeTo(x1-r+k, y0, x1, y0+r-k, x1, y0+r)
	v.LineTo(x1, y1-r)
	v.CubeTo(x1, y1-r+k, x1-r+k, y1, x1-r, y1)
	v.LineTo(x0+r, y1)
	v.CubeTo(x0+r-k, y1, x0, y1-r+k, x0, y1-r)
	v.LineTo(x0, y0+r)
	v.CubeTo(x0, y0+r-k, x0+r-k, y0, x0+r, y0)
	v.ClosePath()
	v.Draw(c.image, c.image.Bounds(), image.NewUniform(ink), image.Point{})
}

func (c *canvas) card(x, y, width, height int) {
	c.rounded(x, y, width, height, cardRadius, c.theme.Border)
	c.rounded(x+1, y+1, width-2, height-2, cardRadius-1, c.theme.Card)
}

func (c *canvas) header(title, subtitle string, paint bool) int {
	titles := wrap(c.title, title, imageWidth-2*margin-gap)
	subtitles := wrap(c.body, subtitle, imageWidth-2*margin)
	y := margin
	if paint {
		c.rounded(margin, y+accentInset, accentWidth, titleLineHeight-2*accentInset, accentRadius, c.theme.Accent)
		c.text(c.title, titles, margin+gap, y, titleLineHeight, c.theme.Text)
	}
	y += len(titles)*titleLineHeight + gap/2
	if subtitle != "" {
		if paint {
			c.text(c.body, subtitles, margin, y, lineHeight, c.theme.Muted)
		}
		y += len(subtitles) * lineHeight
	}
	return y + gap
}

// Index 绘制每行三张卡片的总览，说明最多显示两行。
func Index(f *Fonts, theme Theme, data IndexData) ([]byte, error) {
	c, err := newCanvas(f, theme)
	if err != nil {
		return nil, err
	}
	defer c.close()
	// 1. 按实际字宽计算卡片和各行高度。
	const cardWidth = (imageWidth - 2*margin - (columns-1)*gap) / columns
	const contentWidth = cardWidth - 2*padding
	type cardLayout struct{ name, description []string }
	cards := make([]cardLayout, len(data.Plugins))
	rows := make([]int, (len(cards)+columns-1)/columns)
	for i, p := range data.Plugins {
		cards[i] = cardLayout{wrap(c.bold, p.Name, contentWidth), ellipsis(c.body, wrap(c.body, p.Description, contentWidth), contentWidth, descriptionLines)}
		height := padding + badgeWidth + gap + len(cards[i].name)*boldLineHeight + gap/2 + descriptionLines*lineHeight + gap + lineHeight + padding
		rows[i/columns] = max(rows[i/columns], height)
	}
	start := c.header(data.Title, data.Subtitle, false)
	height := start + margin
	for _, row := range rows {
		height += row + gap
	}
	if len(rows) > 0 {
		height -= gap
	}
	c.start(height)
	c.header(data.Title, data.Subtitle, true)
	// 2. 同行卡片等高，命令数量固定在卡片底部。
	y := start
	for i, p := range data.Plugins {
		x := margin + (i%columns)*(cardWidth+gap)
		h := rows[i/columns]
		c.card(x, y, cardWidth, h)
		badge := strconv.Itoa(i + 1)
		badgeSize := max(badgeWidth, font.MeasureString(c.body, badge).Ceil()+gap)
		c.rounded(x+padding, y+padding, badgeSize, badgeWidth, badgeRadius, c.theme.Badge)
		c.text(c.body, []string{badge}, x+padding+(badgeSize-font.MeasureString(c.body, badge).Ceil())/2, y+padding+(badgeWidth-lineHeight)/2, lineHeight, c.theme.Accent)
		textY := c.text(c.bold, cards[i].name, x+padding, y+padding+badgeWidth+gap, boldLineHeight, c.theme.Text)
		c.text(c.body, cards[i].description, x+padding, textY+gap/2, lineHeight, c.theme.Muted)
		c.text(c.body, wrap(c.body, p.Count, contentWidth), x+padding, y+h-padding-lineHeight, lineHeight, c.theme.Accent)
		if i%columns == columns-1 {
			y += h + gap
		}
	}
	return c.encode()
}

// Detail 绘制插件说明和完整命令列表，所有行随内容增高。
func Detail(f *Fonts, theme Theme, data DetailData) ([]byte, error) {
	c, err := newCanvas(f, theme)
	if err != nil {
		return nil, err
	}
	defer c.close()
	// 1. 计算头部、说明及每条命令的实际高度。
	const contentWidth = imageWidth - 2*margin - 2*padding - badgeWidth - gap
	start := c.header(data.Title, data.Description, false)
	if data.Subtitle != "" {
		start += len(wrap(c.body, data.Subtitle, imageWidth-2*margin))*lineHeight + gap
	}
	heights := make([]int, len(data.Commands))
	height := start + margin
	for i, row := range data.Commands {
		h := 2*padding + len(wrap(c.bold, row.Usage, contentWidth))*boldLineHeight
		for _, text := range []string{row.Aliases, row.Description} {
			if text != "" {
				h += gap/2 + len(wrap(c.body, text, contentWidth))*lineHeight
			}
		}
		heights[i] = h
		height += h + gap
	}
	if len(heights) > 0 {
		height -= gap
	}
	c.start(height)
	y := c.header(data.Title, data.Description, true)
	if data.Subtitle != "" {
		c.text(c.body, wrap(c.body, data.Subtitle, imageWidth-2*margin), margin, y, lineHeight, c.theme.Muted)
	}
	// 2. 左侧序号与右侧用法分栏，长参数、别名与说明独立换行。
	y = start
	for i, row := range data.Commands {
		c.card(margin, y, imageWidth-2*margin, heights[i])
		c.text(c.body, []string{strconv.Itoa(i + 1)}, margin+padding, y+padding, lineHeight, c.theme.Accent)
		x := margin + padding + badgeWidth + gap
		textY := c.text(c.bold, wrap(c.bold, row.Usage, contentWidth), x, y+padding, boldLineHeight, c.theme.Text)
		for _, text := range []string{row.Aliases, row.Description} {
			if text != "" {
				textY = c.text(c.body, wrap(c.body, text, contentWidth), x, textY+gap/2, lineHeight, c.theme.Muted)
			}
		}
		y += heights[i] + gap
	}
	return c.encode()
}

// Command 绘制单个命令的完整用法、别名、说明和所属插件。
func Command(f *Fonts, theme Theme, data CommandData) ([]byte, error) {
	c, err := newCanvas(f, theme)
	if err != nil {
		return nil, err
	}
	defer c.close()
	const contentWidth = imageWidth - 2*margin - 2*padding
	fields := []struct{ label, value string }{
		{data.UsageLabel, data.Usage}, {data.AliasesLabel, data.Aliases}, {data.DescriptionLabel, data.Description}, {data.PluginLabel, data.Plugin},
	}
	start := c.header(data.Title, "", false)
	height := 2 * padding
	for _, field := range fields {
		height += len(wrap(c.bold, field.label, contentWidth))*boldLineHeight + len(wrap(c.body, field.value, contentWidth))*lineHeight + gap
	}
	height -= gap
	c.start(start + height + margin)
	c.header(data.Title, "", true)
	c.card(margin, start, imageWidth-2*margin, height)
	y := start + padding
	for _, field := range fields {
		y = c.text(c.bold, wrap(c.bold, field.label, contentWidth), margin+padding, y, boldLineHeight, c.theme.Accent)
		y = c.text(c.body, wrap(c.body, field.value, contentWidth), margin+padding, y, lineHeight, c.theme.Text) + gap
	}
	return c.encode()
}
