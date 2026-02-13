package tui

import (
	"fmt"
	"strings"

	ansi "github.com/leaanthony/go-ansi-parser"
)

const (
	SVG_FONT_SIZE         = 32.0
	SVG_LINE_HEIGHT_RATIO = 1.2
)

func buildAttrs(obj *ansi.StyledText) string {
	attrs := []string{}
	if obj.Bold() {
		attrs = append(attrs, `font-weight="bold"`)
	}
	if obj.FgCol != nil && obj.FgCol.Hex != "" {
		attrs = append(attrs, fmt.Sprintf(`fill="%s"`, obj.FgCol.Hex))
	}
	if obj.BgCol != nil && obj.BgCol.Hex != "" {
		if strings.TrimSpace(obj.Label) != "" {
			attrs = append(attrs, `text-decoration="underline"`)
		}

	}
	if obj.Faint() {
		attrs = append(attrs, `fill-opacity="0.5"`)
	}
	if len(attrs) == 0 {
		return ""
	}
	return " " + strings.Join(attrs, " ")
}

func ansi2svg(raw string) (string, error) {
	fontSize := SVG_FONT_SIZE
	lineHeight := fontSize * SVG_LINE_HEIGHT_RATIO

	objs, err := ansi.Parse(raw)
	if err != nil {
		return "", err
	}

	header := []string{}
	chunks := []string{}
	maxCols := 0 // max visible characters on any line
	curCols := 0 // visible characters on the current line
	y := lineHeight

	buffer := strings.Builder{}

	for _, obj := range objs {
		if obj.Len == 0 {
			continue
		}

		attrs := buildAttrs(obj)

		if strings.Contains(obj.Label, "\n") {
			lines := strings.Split(obj.Label, "\n")
			buffer.WriteString(lines[0])
			curCols += len(lines[0])
			content := buffer.String()
			chunks = append(chunks, fmt.Sprintf(`<text x="%d" y="%.2f"%s>%s</text>`, 0, y, attrs, content))
			if curCols > maxCols {
				maxCols = curCols
			}
			buffer.Reset()
			curCols = 0

			for _, line := range lines[1 : len(lines)-1] {
				if line == "" {
					continue
				}
				chunks = append(chunks, fmt.Sprintf(`<text x="%d" y="%.2f"%s>%s</text>`, 0, y+lineHeight, attrs, line))
				if len(line) > maxCols {
					maxCols = len(line)
				}
				y += lineHeight
			}

			buffer.WriteString(lines[len(lines)-1])
			curCols = len(lines[len(lines)-1])
			y += lineHeight

		} else {
			tspan := fmt.Sprintf(`<tspan%s>%s</tspan>`, attrs, obj.Label)
			buffer.WriteString(tspan)
			curCols += len(obj.Label)
		}

	}
	if buffer.Len() > 0 {
		if curCols > maxCols {
			maxCols = curCols
		}
		content := buffer.String()
		chunks = append(chunks, fmt.Sprintf(`<text x="%d" y="%.2f">%s</text>`, 0, y, content))
	}

	charWidth := fontSize * 0.2 // empirical ratio for monospace font
	w := float64(maxCols) * charWidth
	header = append(header, fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" xml:space="preserve" viewBox="0 0 %.2f %.2f" font-family="monospace" font-size="%.2f">`, w, y+lineHeight, fontSize))
	header = append(header, `<style>text { white-space: pre; }</style>`)

	all := append(header, chunks...)
	all = append(all, "</svg>")

	return strings.Join(all, "\n"), nil
}
