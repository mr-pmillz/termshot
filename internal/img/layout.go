// Copyright © 2020 The Homeport Team
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package img

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/clipperhouse/uax29/v2/graphemes"
	"github.com/gonvenience/bunt"
	"github.com/mattn/go-runewidth"
)

type contentRange struct {
	start int
	end   int
}

type segmentKind uint8

const (
	segmentText segmentKind = iota
	segmentTab
	segmentNewline
)

type layoutSegment struct {
	Kind        segmentKind
	Text        string
	Settings    uint64
	Width       int
	SourceStart int
	SourceEnd   int
	IsCommand   bool
}

type layoutRow struct {
	Segments   []layoutSegment
	Width      int
	SourceEnd  int
	HasCommand bool
	HasOutput  bool
}

type contentLayout struct {
	Rows    []layoutRow
	MaxCols int
}

func (s *Scaffold) layout() contentLayout {
	return layoutContent(s.content, s.commandRanges, s.wrapColumns(), s.tabSpaces)
}

func (s *Scaffold) wrapColumns() int {
	return s.GetFixedColumns()
}

func layoutContent(content bunt.String, commandRanges []contentRange, columns, tabSpaces int) contentLayout {
	segments := segmentizeContent(content, commandRanges, tabSpaces)
	rows := make([]layoutRow, 0, 8)
	row := layoutRow{}
	lastSourceEnd := 0
	endedWithNewline := false

	flushRow := func(sourceEnd int) {
		row.SourceEnd = sourceEnd
		rows = append(rows, row)
		row = layoutRow{}
	}

	for _, segment := range segments {
		switch segment.Kind {
		case segmentTab, segmentText:
			if columns > 0 && row.Width > 0 && row.Width+segment.Width > columns {
				flushRow(lastSourceEnd)
			}

			row.Segments = append(row.Segments, segment)
			row.Width += segment.Width
			row.HasCommand = row.HasCommand || segment.IsCommand
			row.HasOutput = row.HasOutput || !segment.IsCommand
			lastSourceEnd = segment.SourceEnd
			endedWithNewline = false

		case segmentNewline:
			row.SourceEnd = segment.SourceEnd
			row.HasCommand = row.HasCommand || segment.IsCommand
			row.HasOutput = row.HasOutput || !segment.IsCommand
			rows = append(rows, row)
			row = layoutRow{}
			lastSourceEnd = segment.SourceEnd
			endedWithNewline = true

		default:
			// Reserved for future segment kinds.
		}
	}

	if len(rows) == 0 || !endedWithNewline || len(row.Segments) > 0 || row.HasCommand || row.HasOutput {
		flushRow(lastSourceEnd)
	}

	if endedWithNewline && len(rows) > 1 && len(rows[len(rows)-1].Segments) == 0 && !rows[len(rows)-1].HasCommand && !rows[len(rows)-1].HasOutput {
		rows = rows[:len(rows)-1]
	}

	layout := contentLayout{
		Rows: rows,
	}

	for _, laidOutRow := range rows {
		if laidOutRow.Width > layout.MaxCols {
			layout.MaxCols = laidOutRow.Width
		}
	}

	return layout
}

func segmentizeContent(content bunt.String, commandRanges []contentRange, tabSpaces int) []layoutSegment {
	segments := make([]layoutSegment, 0, len(content))
	for i := 0; i < len(content); {
		current := content[i]

		switch current.Symbol {
		case '\n':
			segments = append(segments, layoutSegment{
				Kind:        segmentNewline,
				Text:        "\n",
				Settings:    current.Settings,
				SourceStart: i,
				SourceEnd:   i + 1,
				IsCommand:   rangeContains(commandRanges, i),
			})
			i++
			continue

		case '\t':
			segments = append(segments, layoutSegment{
				Kind:        segmentTab,
				Text:        "\t",
				Settings:    current.Settings,
				Width:       tabSpaces,
				SourceStart: i,
				SourceEnd:   i + 1,
				IsCommand:   rangeContains(commandRanges, i),
			})
			i++
			continue
		}

		settings := current.Settings
		isCommand := rangeContains(commandRanges, i)
		start := i
		var builder strings.Builder
		for i < len(content) {
			r := content[i]
			if r.Symbol == '\n' || r.Symbol == '\t' || r.Settings != settings || rangeContains(commandRanges, i) != isCommand {
				break
			}

			builder.WriteRune(r.Symbol)
			i++
		}

		run := builder.String()
		position := start
		tokens := graphemes.FromString(run)
		for tokens.Next() {
			cluster := tokens.Value()
			clusterLen := utf8.RuneCountInString(cluster)
			segments = append(segments, layoutSegment{
				Kind:        segmentText,
				Text:        cluster,
				Settings:    settings,
				Width:       runewidth.StringWidth(cluster),
				SourceStart: position,
				SourceEnd:   position + clusterLen,
				IsCommand:   isCommand,
			})
			position += clusterLen
		}
	}

	return segments
}

func rangeContains(ranges []contentRange, index int) bool {
	for _, current := range ranges {
		if index < current.start {
			return false
		}
		if index < current.end {
			return true
		}
	}

	return false
}

func plainContent(text string) bunt.String {
	result := make(bunt.String, 0, len(text))
	for _, r := range text {
		result = append(result, bunt.ColoredRune{Symbol: r})
	}

	return result
}

func trimCommandRanges(ranges []contentRange, limit int) []contentRange {
	trimmed := make([]contentRange, 0, len(ranges))
	for _, current := range ranges {
		if current.start >= limit {
			break
		}

		if current.end > limit {
			current.end = limit
		}

		trimmed = append(trimmed, current)
	}

	return trimmed
}

func (s *Scaffold) truncationFooter(omittedRows int) string {
	wrapCols := s.wrapColumns()
	full := fmt.Sprintf("... [%d rows truncated] ...", omittedRows)
	if wrapCols <= 0 {
		return full
	}

	candidates := []string{
		full,
		fmt.Sprintf("... [%d] ...", omittedRows),
		fmt.Sprintf("[%d]", omittedRows),
		"...",
		".",
	}

	for _, candidate := range candidates {
		if runewidth.StringWidth(candidate) <= wrapCols {
			return candidate
		}
	}

	return "."
}

// LimitRenderedRows truncates non-command content to at most maxRows rendered
// rows after wrapping. If truncation occurs, a footer is appended inside the
// same row budget.
func (s *Scaffold) LimitRenderedRows(maxRows int) {
	if maxRows <= 0 {
		return
	}

	layout := s.layout()
	outputRows := 0
	for _, row := range layout.Rows {
		if row.HasOutput {
			outputRows++
		}
	}

	if outputRows <= maxRows {
		return
	}

	rowsForContent := maxRows
	if maxRows > 0 {
		rowsForContent = maxRows - 1
		if rowsForContent < 0 {
			rowsForContent = 0
		}
	}

	keptOutputRows := 0
	keepSourceEnd := 0
	for _, row := range layout.Rows {
		if !row.HasOutput {
			if row.SourceEnd > keepSourceEnd {
				keepSourceEnd = row.SourceEnd
			}
			continue
		}

		if keptOutputRows >= rowsForContent {
			break
		}

		keepSourceEnd = row.SourceEnd
		keptOutputRows++
	}

	omittedRows := outputRows - keptOutputRows
	footer := s.truncationFooter(omittedRows)
	newContent := append(bunt.String{}, s.content[:keepSourceEnd]...)
	if len(newContent) > 0 && newContent[len(newContent)-1].Symbol != '\n' {
		newContent = append(newContent, bunt.ColoredRune{Symbol: '\n'})
	}
	newContent = append(newContent, plainContent(footer)...)

	s.content = newContent
	s.commandRanges = trimCommandRanges(s.commandRanges, keepSourceEnd)
}
