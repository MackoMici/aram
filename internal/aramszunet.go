//Felolvassa az aramszunet.txt file-t és szétbontja majd csinál egy indexet belőle varos-utca-házszám szintjén, hogy később lehessen keresni benne

package internal

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/MackoMici/aram/config"
	"github.com/MackoMici/aram/logging"
)

type AramSzunet struct {
	ID           string
	Datum        string
	Idoszak      string
	Varos        string
	VarosLink    string
	Terulet      string
	Hazszamok    []int
	Megjegyzes   string
	Teruletgazda string
	Forras       string
	Bekerules    string
	Terulet_mod  string
	TeljesTel    bool
}

type AramSzunets struct {
	List  []*AramSzunet
	file  string
	index map[string]map[string]map[int][]*AramSzunet
}

var aram_patterns []*regexp.Regexp
var hazszam_patterns []*regexp.Regexp
var clean_patterns []*regexp.Regexp
var kizar_patterns []*regexp.Regexp
var aram_replace []*config.Replacements

func NewAramSzunets(file string, conf *config.Config) *AramSzunets {
	am := &AramSzunets{
		file: file,
	}

	for _, p := range conf.AramszunetPatterns {
		re, err := regexp.Compile(p)
		if err != nil {
			logging.Logger.Error("Érvénytelen pattern ", p, err)
		}
		aram_patterns = append(aram_patterns, re)
	}
	for _, p := range conf.HazszamPatterns {
		re, err := regexp.Compile(p)
		if err != nil {
			logging.Logger.Error("Érvénytelen pattern ", p, err)
		}
		hazszam_patterns = append(hazszam_patterns, re)
	}
	for _, p := range conf.KizarPatterns {
		re, err := regexp.Compile(p)
		if err != nil {
			logging.Logger.Error("Érvénytelen pattern ", p, err)
		}
		kizar_patterns = append(kizar_patterns, re)
	}
	for _, p := range conf.CleanPatterns {
		re, err := regexp.Compile(p)
		if err != nil {
			logging.Logger.Error("Érvénytelen pattern ", p, err)
		}
		clean_patterns = append(clean_patterns, re)
	}
	for _, p := range conf.AramszunetReplacements {
		aram_replace = append(aram_replace, p)
	}
	am.Load()
	am.BuildIndex()
	return am
}

func (a *AramSzunets) Load() {
	// open file
	f, err := os.Open(a.file)
	if err != nil {
		logging.Fatal("Áramszünet fájl megnyitás", "hiba", err)
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
			logging.Logger.Error("ERROR", "error", err)
			break
		}
		a.List = append(a.List, NewAramSzunet(rStr))
	}
	logging.Logger.Info("Áramszünet", "darab", len(a.List))
}

func (a *AramSzunets) BuildIndex() {
	a.index = make(map[string]map[string]map[int][]*AramSzunet)
	hrszSeen := make(map[string]bool) // városonkénti hrsz nyilvántartás
	for _, rec := range a.List {
		city := rec.Varos
		street := rec.Terulet_mod
		dateKey := city + "|" + rec.Datum

		if strings.Contains(strings.ToLower(street), "hrsz") {
			if hrszSeen[dateKey] {
				continue
			}
			hrszSeen[dateKey] = true
		}

		if len(rec.Hazszamok) == 0 {
			rec.Hazszamok = []int{0} // ha nincs házszám, akkor legyen 0
		}
		if a.index[city] == nil {
			a.index[city] = make(map[string]map[int][]*AramSzunet)
		}
		if a.index[city][street] == nil {
			a.index[city][street] = make(map[int][]*AramSzunet)
		}

		for _, num := range rec.Hazszamok {
			a.index[city][street][num] = append(a.index[city][street][num], rec)
		}
	}
	logging.Logger.Debug("Áramszünet index", "Index", a.index)
}

func (a *AramSzunets) Find(city, street string, number int) []*AramSzunet {
	if a.index == nil {
		a.BuildIndex()
	}
	if cityMap, ok := a.index[city]; ok {
		if streetMap, ok := cityMap[street]; ok {
			return streetMap[number]
		}
	}
	return nil
}

func NewAramSzunet(data []string) *AramSzunet {
	a := &AramSzunet{
		ID:         data[0],
		Datum:      data[1],
		Idoszak:    data[2],
		Varos:      teruletMod(data[3]),
		VarosLink:  data[4],
		Terulet:    data[5],
		Megjegyzes: data[6],
		Forras:     data[7],
		Bekerules:  data[8],
	}
	a.setTerulet(data[5])
	return a
}

func (a *AramSzunet) setTerulet(s string) {
	re := regexp.MustCompile(`Teljes település`)
	a.TeljesTel = re.MatchString(s)
	for _, p := range aram_patterns {
		parts := p.FindStringSubmatch(s)
		streetRaw, numsRaw := parts[1], parts[3]
		a.Terulet_mod = teruletMod(streetRaw)
		if numsRaw != "" {
			nums, err := parseHouseNumbers(numsRaw)
			if err != nil {
				logging.Logger.Error("Házszám parse hiba", numsRaw, err)
			} else {
				a.Hazszamok = nums
			}
		}
	}
}

func parseHouseNumbers(s string) ([]int, error) {
	var nums []int
	var tokens []string

	for _, re := range hazszam_patterns {
		tokens = re.FindAllString(s, -1)
	}
	for i := 0; i < len(tokens); i++ {
		tok := strings.TrimSpace(tokens[i])
		tok = cleanPatterns(tok)
		if isIgnored(tok) {
			continue // kihagyjuk
		}
		if strings.Contains(tok, " - ") {
			ends := strings.SplitN(tok, " - ", 2)
			start, err := toInt(ends[0])
			if err != nil {
				return nil, fmt.Errorf("érvénytelen tartomány kezdete %q: %w", ends[0], err)
			}
			end, err := toInt(ends[1])
			if err != nil {
				return nil, fmt.Errorf("érvénytelen tartomány vége %q: %w", ends[1], err)
			}
			if end < start {
				start, end = end, start
			}
			for n := start; n <= end; n += 2 {
				nums = append(nums, n)
			}
			continue
		}
		n, err := toInt(tok)
		if err != nil {
			return nil, fmt.Errorf("érvénytelen szám %q: %w", tok, err)
		}
		nums = append(nums, n)
	}
	sort.Ints(nums)
	out := nums[:0]
	for i, v := range nums {
		if i == 0 || v != nums[i-1] {
			out = append(out, v)
		}
	}
	return out, nil
}

func toInt(str string) (int, error) {
	clean := strings.TrimRight(str, ".")
	if idx := strings.IndexAny(clean, "/ "); idx != -1 {
		clean = clean[:idx]
	}
	return strconv.Atoi(clean)
}

func cleanPatterns(token string) string {
	for _, re := range clean_patterns {
		token = re.ReplaceAllString(token, "")
	}
	return token
}

func isIgnored(token string) bool {
	for _, re := range kizar_patterns {
		if re.MatchString(token) {
			return true
		}
	}
	return false
}

func teruletMod(s string) string {
	if len(aram_replace) > 0 {
		for _, r := range aram_replace {
			s = r.Replace(s)
		}
	}
	return s
}

func (a *AramSzunet) String() string {
	return fmt.Sprintf("%s %s - %s; %s %s", a.ID, a.Datum, a.Idoszak, a.Varos, a.Terulet)
}
