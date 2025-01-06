package internal

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/MackoMici/aram/config"
)

type Node struct {
	Irszam       string
	Varos        string
	Node         string
        Terulet      string
	nfo          string
        Vegpont_mod1 string
	Vegpont_mod2 string
	Sarok        bool
}

type Nodes struct {
	List     []*Node
	file     string
	vegponts map[string]*Node
}

var node_patterns []*regexp.Regexp

func NewNodes(file string, conf *config.Config) *Nodes {
	am := &Nodes{
		file:     file,
		vegponts: make(map[string]*Node),
	}
	for _, p := range conf.TeruletPatterns {
		re, err := regexp.Compile(p)
		if err != nil {
			log.Println("Invalid pattern ", p, err)
		}
		node_patterns = append(node_patterns, re)
	}
	am.Load()
	return am
}

func (a *Nodes) Load() {
	// open file
	f, err := os.Open(a.file)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	fcsv := csv.NewReader(f)
	fcsv.Comma = '\t'
	for {
		rStr, err := fcsv.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Println("ERROR: ", err.Error())
			break
		}
		am := NewNode(rStr)
		a.List = append(a.List, am)
		if am.Sarok {
			a.vegponts[am.Vegpont_mod2] = am
		}
		a.vegponts[am.Vegpont_mod1] = am
	}
	log.Println("Node darabszám: ", len(a.List))
}

func (a *Nodes) Vegpont(vegpont string) *Node {
	if v, ok := a.vegponts[vegpont]; ok {
		return v
	}
	return nil
}

func NewNode(data []string) *Node {
	a := &Node{
		Irszam:  data[0],
		Varos:   data[1],
                Node:    data[2],
		Terulet: data[3],
                nfo:     data[4],
	}
	a.setVegpont(data[3])
	return a
}

func (a *Node) setVegpont(s string) {
	re := regexp.MustCompile(` sarok`)
        a.Sarok = re.MatchString(s)
	for _, p := range node_patterns {
		if a.Sarok {
			r := strings.Split(s, " - ")
			a.Vegpont_mod2 = fmt.Sprintf("%s %s", a.Varos, p.ReplaceAllString(r[1], ""))
			a.Vegpont_mod1 = fmt.Sprintf("%s %s", a.Varos, p.ReplaceAllString(r[0], ""))
		} else {
			a.Vegpont_mod1 = fmt.Sprintf("%s %s", a.Varos, p.ReplaceAllString(s, ""))
		}
		break
	}
}

func (a *Node) String() string {
	return fmt.Sprintf("%s node, %s %s", a.Node, a.Varos, a.Terulet)
}
