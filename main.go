package main

import (
	"compress/bzip2"
	"io"
	"log"
	"os"
	"strings"

	"github.com/dustin/go-wikiparse"
)

type Page struct {
	Title string
	Runes map[rune]uint32
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
	for {
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
		p := &Page{Title: page.Title, Runes: map[rune]uint32{}}
		total := 0
		for _, codepoint := range page.Revisions[0].Text {
			p.Runes[codepoint]++
			total++
		}
		log.Printf("Length=%d, Unique runes=%d L/r=%f", total, len(p.Runes), float64(total)/float64(len(p.Runes)))
		pages[p.Title] = p
	}
	return nil
}

func main() {
	err := WorkWIthDumpFile(`dump.bz2`)
	if err != nil {
		log.Fatal(err)
	}
}
