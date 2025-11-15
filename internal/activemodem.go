//megnézi, hogy az adott node-ban lévő címeken hol lesz áramszünet és összeszámolja azt

package internal

import (
	"encoding/csv"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/MackoMici/aram/config"
	"github.com/MackoMici/aram/logging"
)

type ActiveModem struct {
	Node    string
	Vegpont string
	ID      string
	Node1   string
	Node2   string
	Varos   string
	Terulet string
	Hazszam int
}

type ActiveModems struct {
	List  []*ActiveModem
	file  string
	index map[string]map[string]map[int][]*ActiveModem
}

var vegpont_patterns []*regexp.Regexp

func NewActiveModems(file string, conf *config.Config) *ActiveModems {
	am := &ActiveModems{
		file:  file,
		index: make(map[string]map[string]map[int][]*ActiveModem),
	}

	for _, p := range conf.VegpontPatterns {
		re, err := regexp.Compile(p)
		if err != nil {
			logging.Logger.Error("Érvénytelen pattern ", p, err)
		}
		vegpont_patterns = append(vegpont_patterns, re)
	}
	am.Load()
	return am
}

func (a *ActiveModems) Load() {
	// open file
	f, err := os.Open(a.file)
	if err != nil {
		logging.Fatal("ActiveModem fájl megnyitás", "hiba", err)
	}
	defer f.Close()
	fcsv := csv.NewReader(f)
	fcsv.Comma = ';'
	fcsv.LazyQuotes = true
	for {
		rStr, err := fcsv.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			logging.Logger.Error("ERROR", "error", err)
			break
		}
		a.List = append(a.List, NewActiveModem(rStr))
	}
	logging.Logger.Info("ActiveModem", "darab", len(a.List))
}

func NewActiveModem(data []string) *ActiveModem {
	a := &ActiveModem{
		Node:    data[13],
		Vegpont: data[14],
		ID:      data[15],
	}
	a.setNode(data[13])
	a.setVegpont(data[14])
	return a
}
func (a *ActiveModem) setNode(s string) {
	if strings.Contains(s, ";") {
		r := strings.Split(s, ";")
		a.Node1 = r[0]
		a.Node2 = r[1]
	} else {
		a.Node1 = s
	}
}

func (a *ActiveModem) setVegpont(s string) {
	for _, p := range vegpont_patterns {
		if namedGroups := a.matchWithGroup(p, s); len(namedGroups) > 0 {
			a.Varos = a.varos(namedGroups)
			a.Terulet = a.terulet(namedGroups)
			if a.hazszam(namedGroups) == "" {
				a.Hazszam = -1
			}
			i, err := strconv.Atoi(a.hazszam(namedGroups))
			if err != nil {
				logging.Logger.Error("ActiveModem", a.hazszam(namedGroups), err.Error())
			}
			a.Hazszam = i
		}
	}
}

func (a *ActiveModem) varos(namedgroups map[string]string) string {
	s, ok := namedgroups["Varos"]
	if !ok {
		return ""
	}
	return s
}

func (a *ActiveModem) terulet(namedgroups map[string]string) string {
	s, ok := namedgroups["Terulet"]
	if !ok {
		return ""
	}
	return s
}

func (a *ActiveModem) hazszam(namedgroups map[string]string) string {
	s, ok := namedgroups["Hazszam"]
	if !ok {
		return ""
	}
	return s
}

func (a *ActiveModem) matchWithGroup(r *regexp.Regexp, s string) map[string]string {
	namedGroups := make(map[string]string)
	if match := r.FindStringSubmatch(s); len(match) > 0 {
		for i, name := range r.SubexpNames() {
			if i != 0 && name != "" {
				namedGroups[name] = match[i]
			}
		}
	}
	return namedGroups
}

func (a *ActiveModems) FindByNode(nodeName string) []*ActiveModem {
	var result []*ActiveModem
	for _, modem := range a.List {
		if modem.Node1 == nodeName || modem.Node2 == nodeName {
			result = append(result, modem)
			logging.Logger.Debug("FindByNode", "node", nodeName, "végpont", modem.Vegpont)
		}
	}
	if len(result) == 0 {
		logging.Logger.Debug("FindByNode", nodeName, "nincs végpont")
	}
	return result
}

func FilterAffectedModems(modems []*ActiveModem, outages *AramSzunets) []*ActiveModem {
	var affected []*ActiveModem
	for _, modem := range modems {
		if matches := outages.Find(modem.Varos, modem.Terulet, modem.Hazszam); matches != nil {
			logging.Logger.Debug("FilterAffectedModems", "Város", modem.Varos, "Terület", modem.Terulet, "Házszám", modem.Hazszam)
			affected = append(affected, modem)
		}
	}
	return affected
}
