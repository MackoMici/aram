//megnézi, hogy az adott node-ban lévő címeken hol lesz áramszünet és összeszámolja azt

package internal

import (
	"encoding/csv"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/MackoMici/aram/config"
	"github.com/MackoMici/aram/logging"
)

type ActiveModem struct {
	Node    string
	Vegpont string
	ID      string
	Node1   string
	Node2   string
	Node3   string
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
	ApplyCityOverrides(am, conf)
	return am
}

func ApplyCityOverrides(am *ActiveModems, conf *config.Config) {
	logged := make(map[string]bool)
	for _, modem := range am.List {
		// Node alapján város név csere
		if newCity, ok := conf.AmVarosReplacements[modem.Node1]; ok {
			if !logged[modem.Node1] {
				logging.Logger.Debug("AmVarosReplacements", "node1", modem.Node1, "régi város", modem.Varos, "új város", newCity)
				logged[modem.Node1] = true
			}
			modem.Varos = newCity
		}
		if newCity, ok := conf.AmVarosReplacements[modem.Node2]; ok {
			if !logged[modem.Node2] {
				logging.Logger.Debug("AmVarosReplacements", "node2", modem.Node2, "régi város", modem.Varos, "új város", newCity)
				logged[modem.Node2] = true
			}
			modem.Varos = newCity
		}
		if newCity, ok := conf.AmVarosReplacements[modem.Node3]; ok {
			if !logged[modem.Node3] {
				logging.Logger.Debug("AmVarosReplacements", "node3", modem.Node3, "régi város", modem.Varos, "új város", newCity)
				logged[modem.Node3] = true
			}
			modem.Varos = newCity
		}
		// Utca alapján város és utca név csere
		parts := strings.Fields(modem.Terulet)
		if len(parts) > 1 {
			if newStreet, ok := conf.AmUtcaReplacements[parts[0]]; ok {
				if !logged[modem.Terulet] {
					logging.Logger.Debug("AmUtcaReplacements", "régi utca", modem.Terulet, "új utca", strings.Join(parts[1:], " "), "régi város", modem.Varos, "új város", newStreet)
					logged[modem.Terulet] = true
				}
				modem.Terulet = strings.Join(parts[1:], " ")
				modem.Varos = newStreet
			}
		}
	}
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
		if len(r) > 2 {
			a.Node3 = r[2]
		}
	} else {
		a.Node1 = s
	}
}

func (a *ActiveModem) setVegpont(s string) {
	for _, p := range vegpont_patterns {
		if namedGroups := a.matchWithGroup(p, s); len(namedGroups) > 0 {
			a.Varos = a.varos(namedGroups)
			a.Terulet = strings.ReplaceAll(a.terulet(namedGroups), "  ", " ")
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
		if modem.Node1 == nodeName || modem.Node2 == nodeName || modem.Node3 == nodeName {
			result = append(result, modem)
			logging.Logger.Debug("FindByNode", "node", nodeName, "végpont", modem.Vegpont)
		}
	}
	if len(result) == 0 {
		logging.Logger.Error("FindByNode", nodeName, "nincs végpont")
	}
	return result
}

func FilterAffectedModems(modems []*ActiveModem, outages *AramSzunets, datum time.Time) []*ActiveModem {
	var affected []*ActiveModem
	for _, modem := range modems {
		if matches := outages.Find(modem.Varos, modem.Terulet, modem.Hazszam); matches != nil {
			// dátum szűrés
			for _, outage := range matches {
				parsedOutageDate, err := time.Parse("2006-01-02", outage.Datum)
				if err != nil {
					// ha rossz formátum, logoljuk
					logging.Logger.Error("Hibás Dátum", outage.Datum, err)
					continue
				}
				if parsedOutageDate.Equal(datum) {
					logging.Logger.Debug("FilterAffectedModems", "Város", modem.Varos, "Terület", modem.Terulet, "Házszám", modem.Hazszam)
					affected = append(affected, modem)
					break // ha egy cím érintett az adott napon, már hozzáadjuk
				}
			}
		}
	}
	return affected
}
