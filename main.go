package main

import (
	"compress/bzip2"
	"flag"
	"io"
	"log"
	"os"
	"runtime"
	"strings"
	"unicode"

	"github.com/dustin/go-wikiparse"
)

type Page struct {
	Title string
	//	Runes map[rune]uint16
	//	Runes []RuneCount
	runes []runeCount2
}

type runeCount2 struct {
	rune  uint16
	count uint16
}
type runeCount struct {
	rune  rune
	count uint32
}

type PageProcessor interface {
	Process(*wikiparse.Page) error
}

func WorkWIthDumpFile(filename string, work PageProcessor) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	var rd io.Reader = f
	if strings.HasSuffix(filename, `bz2`) {
		rd = bzip2.NewReader(f)
	}
	p, err := wikiparse.NewParser(rd)
	if err != nil {
		return err
	}
	return WorkWIthParser(p, work)
}

func satu16(u uint32) uint16 {
	if u > 0xFFFF {
		return 0xFFFF
	}
	return uint16(u)
}

type miscStats struct {
	pages              map[string]*Page
	totalRuneLengthSum uint64
}

func (ms *miscStats) Process(page *wikiparse.Page) error {
	rmap := map[rune]uint32{}
	p := &Page{Title: string([]byte(page.Title))}
	total := 0
	for _, codepoint := range page.Revisions[0].Text {
		if codepoint < 0x80 || !unicode.Is(unicode.Han, codepoint) {
			continue
		}
		if codepoint > 0xFFFF {
			log.Printf("Rare codepoint 0x%X  = %d '%c'", codepoint, codepoint, codepoint)
		}
		rmap[codepoint]++
		total++
		if len(rmap) > 1500 {
			log.Println("SKIP too many unique codepoints")
			return nil
		}
	}
	//		p.Runes = rmap
	p.runes = make([]runeCount2, len(rmap))
	i := 0
	for k, v := range rmap {
		p.runes[i] = runeCount2{satu16(uint32(k)), satu16(v)}
	}
	ms.totalRuneLengthSum += uint64(len(p.runes))
	log.Printf("Length=%d, Unique runes=%d L/r=%f", total, len(p.runes), float64(total)/float64(len(p.runes)))
	ms.pages[p.Title] = p
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	lpages := uint64(len(ms.pages))
	log.Printf("HeapAlloc=%d HeapObjects=%d npages=%d bytes/page=%d runes/page=%d", mem.HeapAlloc, mem.HeapObjects, lpages, mem.HeapAlloc/lpages, ms.totalRuneLengthSum/lpages)
	return nil
}

func WorkWIthParser(parser wikiparse.Parser, work PageProcessor) error {
	si := parser.SiteInfo()
	log.Println(si.SiteName, si.Base)
	infinite := *maxread <= 0
	for i := 0; infinite || i < *maxread; i++ {
		page, err := parser.Next()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		log.Printf("PAGE %d \"%s\" id=%d ns=%d nrevs=%d", i, page.Title, page.ID, page.Ns, len(page.Revisions))
		if page.Redir.Title != "" {
			log.Println("SKIP redirection ->", page.Redir)
			continue
		}
		if page.Ns != 0 {
			log.Println("SKIP nonzero namespace")
			continue
		}
		if len(page.Revisions) != 1 {
			log.Println("SKIP irregular number of revisions")
			continue
		}
		err = work.Process(page)
		if err != nil {
			return err
		}
	}
	return nil
}

var maxread = flag.Int("maxread", -1, "Maximum number of articles to read")

func main() {
	flag.Parse()
	//	err := WorkWIthDumpFile(`dump.bz2`, &miscStats{map[string]*Page{}, 0})
	//	w := &popularHan{map[rune]uint32{}}
	w := NewScoreByHanzis(ScoreByHanziConfig{Known: 900, Learning: 100, MaxUnknown: 25})
	err := WorkWIthDumpFile(`dump.bz2`, w)
	if err != nil {
		log.Printf("ERROR: %T %v", err, err)
	}
	//	log.Println(len(w.m))
	//	w.Print(os.Stdout)
}
