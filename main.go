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
	//Runes map[rune]uint16
	Runes []RuneCount
}

type RuneCount struct {
	Rune  rune
	Count uint32
}

func WorkWIthDumpFile(filename string) error {
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
	return WorkWIthParser(p)
}

func WorkWIthParser(parser wikiparse.Parser) error {
	si := parser.SiteInfo()
	log.Println(si.SiteName, si.Base)
	pages := map[string]*Page{}
	infinite := *maxread <= 0
	for i := 0; infinite || i < *maxread; i++ {
		page, err := parser.Next()
		if err != nil {
			log.Println("ERROR", page, err)
			return err
		}
		log.Printf("%s id=%d ns=%d nrevs=%d", page.Title, page.ID, page.Ns, len(page.Revisions))
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
		rmap := map[rune]uint32{}
		p := &Page{Title: page.Title}
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
		}
		//		p.Runes = rmap
		p.Runes = make([]RuneCount, 0, len(rmap))
		for k, v := range rmap {
			p.Runes = append(p.Runes, RuneCount{k, v})
		}
		log.Printf("Length=%d, Unique runes=%d L/r=%f", total, len(p.Runes), float64(total)/float64(len(p.Runes)))
		pages[p.Title] = p
		var ms runtime.MemStats
		runtime.ReadMemStats(&ms)
		log.Printf("HeapAlloc=%d HeapObjects=%d npages=%d bytes/page=%d", ms.HeapAlloc, ms.HeapObjects, len(pages), ms.HeapAlloc/uint64(len(pages)))
	}
	_ = pages
	return nil
}

var maxread = flag.Int("maxread", -1, "Maximum number of articles to read")

func main() {
	flag.Parse()
	err := WorkWIthDumpFile(`dump.bz2`)
	if err != nil {
		log.Fatal(err)
	}
}
