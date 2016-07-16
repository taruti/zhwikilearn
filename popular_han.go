package main

import (
	"fmt"
	"io"
	"sort"
	"unicode"

	"github.com/dustin/go-wikiparse"
)

type popularHan struct {
	m map[rune]uint32
}

func (ph *popularHan) Process(page *wikiparse.Page) error {
	for _, codepoint := range page.Revisions[0].Text {
		if codepoint < 0x80 || !unicode.Is(unicode.Han, codepoint) {
			continue
		}
		old := ph.m[codepoint]
		if old != 0xFFFFFFFF {
			ph.m[codepoint] = old + 1
		}
	}
	return nil
}

type runeCountPopularity []runeCount

func (arr runeCountPopularity) Len() int           { return len(arr) }
func (arr runeCountPopularity) Less(i, j int) bool { return arr[i].count > arr[j].count }
func (arr runeCountPopularity) Swap(i, j int)      { arr[i], arr[j] = arr[j], arr[i] }

func (ph *popularHan) Print(w io.Writer) error {
	arr := make([]runeCount, 0, len(ph.m))
	for k, v := range ph.m {
		arr = append(arr, runeCount{k, v})
	}
	sort.Sort(runeCountPopularity(arr))
	for i, v := range arr {
		fmt.Fprintf(w, "%5d %04X %c %d\n", i, v.rune, v.rune, v.count)
	}
	return nil
}

// Only one occurence per page counted
type popularHanByPage popularHan

func (ph *popularHanByPage) Process(page *wikiparse.Page) error {
	tmp := map[rune]struct{}{}
	for _, codepoint := range page.Revisions[0].Text {
		if codepoint < 0x80 || !unicode.Is(unicode.Han, codepoint) {
			continue
		}
		tmp[codepoint] = struct{}{}
	}
	for codepoint, _ := range tmp {
		old := ph.m[codepoint]
		if old != 0xFFFFFFFF {
			ph.m[codepoint] = old + 1
		}
	}
	return nil
}

func (ph *popularHanByPage) Print(w io.Writer) error {
	return (*popularHan)(ph).Print(w)
}
