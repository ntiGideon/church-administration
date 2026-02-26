package main

import (
	"encoding/json"
	"fmt"
	"github.com/ntiGideon/ui"
	"html/template"
	"io/fs"
	"math"
	"net/url"
	"path/filepath"
	"strings"
	"time"
)

func newTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	// Regular pages
	pages, err := fs.Glob(ui.Files, "html/pages/*.gohtml")
	if err != nil {
		return nil, err
	}

	// Process regular pages
	for _, page := range pages {
		name := filepath.Base(page)

		patterns := []string{
			"html/base.gohtml",
			"html/partials/*.gohtml",
			page,
		}

		ts, err := template.New(name).Funcs(functions).ParseFS(ui.Files, patterns...)
		if err != nil {
			return nil, err
		}

		cache[name] = ts
	}

	return cache, nil
}

func formatMoney(amount float64) string {
	abs := math.Abs(amount)
	sign := ""
	if amount < 0 {
		sign = "-"
	}
	switch {
	case abs >= 1_000_000_000:
		return fmt.Sprintf("%s%.2fB", sign, abs/1_000_000_000)
	case abs >= 1_000_000:
		return fmt.Sprintf("%s%.2fM", sign, abs/1_000_000)
	case abs >= 1_000:
		return fmt.Sprintf("%s%.2fK", sign, abs/1_000)
	default:
		return fmt.Sprintf("%s%.2f", sign, abs)
	}
}

func formatDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("Jan 2, 2006")
}

func removeQueryParam(param, urlStr string) string {
	u, err := url.Parse(urlStr)
	if err != nil {
		return "/posts"
	}

	q := u.Query()
	q.Del(param)
	u.RawQuery = q.Encode()

	return u.String()
}

func initials(name string) string {
	if name == "" {
		return "U"
	}

	words := strings.Fields(name)
	if len(words) >= 2 {
		return string(words[0][0]) + string(words[1][0])
	}
	return string(name[0])
}

func toJSON(v interface{}) string {
	a, _ := json.Marshal(v)
	return string(a)
}

var functions = template.FuncMap{
	"formatDate":       formatDate,
	"formatMoney":      formatMoney,
	"multiply":         func(a, b int) int { return a * b },
	"multiplyF":        func(a, b float64) float64 { return a * b },
	"initials":         initials,
	"removeQueryParam": removeQueryParam,
	"json":             toJSON,
	"join":             strings.Join,
	"split":            strings.Split,
	"toLower":          strings.ToLower,
	"toUpper":          strings.ToUpper,
	"title":            strings.Title,
	"isPositive": func(f float64) bool {
		return f >= 0
	}, "isGT5": func(f float64) bool {
		return f >= 5
	}, "isGT2": func(f float64) bool {
		return f >= 2
	},
	"abs":      math.Abs,
	"subtract": func(a, b float64) float64 { return a - b },
	"add":      func(a, b float64) float64 { return a + b },
	"float64":  func(i int) float64 { return float64(i) },
	"int":      func(f float64) int { return int(f) },
}
