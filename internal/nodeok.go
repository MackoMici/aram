package internal

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"strconv"

	"github.com/MackoMici/aram/config"
	"github.com/MackoMici/aram/logging"
)

type Node struct {
	Irszam       string
	Varos        string
	Node         string
	Terulet      string
	Hazszam      int
	nfo          string
	Vegpont_mod1 string
	Vegpont_mod2 string
	Sarok        bool
}

type Nodes struct {
	List  []*Node
	file  string
	index map[string]map[string]map[int][]*Node
}

var node_patterns []*regexp.Regexp

func NewNodes(file string, conf *config.Config) *Nodes {
	am := &Nodes{
		file:  file,
		index: make(map[string]map[string]map[int][]*Node),
	}
	for _, p := range conf.TeruletPatterns {
		re, err := regexp.Compile(p)
		if err != nil {
			logging.Logger.Error("Érvénytelen pattern ", p, err)
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
		logging.Fatal("Nodeok fájl megnyitás", "hiba", err)
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
			logging.Logger.Error("Node", "Load", err.Error())
			break
		}
		am := NewNode(rStr)
		a.List = append(a.List, am)
	}
	logging.Logger.Info("Node", "darab", len(a.List))
	a.BuildIndex()
}

func (a *Nodes) BuildIndex() {
	a.index = make(map[string]map[string]map[int][]*Node)
	for _, node := range a.List {
		city, street, num := node.Varos, node.Vegpont_mod1, node.Hazszam
		// város tábla init
		if a.index[city] == nil {
			a.index[city] = make(map[string]map[int][]*Node)
		}
		// utca tábla init
		if a.index[city][street] == nil {
			a.index[city][street] = make(map[int][]*Node)
		}
		// indexelt házszám → Node
		a.index[city][street][num] = append(a.index[city][street][num], node)
		if node.Sarok {
			street2 := node.Vegpont_mod2
			if a.index[city][street2] == nil {
				a.index[city][street2] = make(map[int][]*Node)
			}
			a.index[city][street2][num] = append(a.index[city][street2][num], node)
		}
	}
	logging.Logger.Debug("Node index", "lista", a.index)
}

func (a *Nodes) Find(city, street string, number int) []*Node {
	if a.index == nil {
		a.BuildIndex()
	}
	if cityMap, ok := a.index[city]; ok {
		if streetMap, ok := cityMap[street]; ok {
			if nodes, ok := streetMap[number]; ok {
				return nodes
			}
			var sarokNodes []*Node
			for _, nodeList := range streetMap {
				for _, node := range nodeList {
					if node.Sarok {
						sarokNodes = append(sarokNodes, node)
					}
				}
			}
			if len(sarokNodes) > 0 {
				return sarokNodes
			}

		}
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
		parts := p.FindStringSubmatch(s)
		streetRaw, numRaw := parts[1], parts[2]
		if a.Sarok {
			r := strings.Split(streetRaw, " - ")
			a.Vegpont_mod2 = r[1]
			a.Vegpont_mod1 = r[0]
		} else {
			a.Vegpont_mod1 = streetRaw
			if numRaw != "" {
				i, err := strconv.Atoi(numRaw)
				if err != nil {
					logging.Logger.Error("Node", numRaw, err.Error())
				}
				a.Hazszam = i
			}
		}
		break
	}
}

func (a *Node) String() string {
	return fmt.Sprintf("%s node, %s %s", a.Node, a.Varos, a.Terulet)
}
