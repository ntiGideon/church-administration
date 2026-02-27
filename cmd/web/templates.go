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
	"enumStr":  func(v interface{ String() string }) string { return v.String() },
	"docCategoryLabel": func(s string) string {
		switch s {
		case "minutes":
			return "Meeting Minutes"
		case "bulletin":
			return "Bulletin"
		case "constitution":
			return "Constitution / Bylaws"
		case "form":
			return "Form"
		case "report":
			return "Report"
		case "financial":
			return "Financial Statement"
		case "other":
			return "Other"
		default:
			return s
		}
	},
	"docCategoryIcon": func(s string) string {
		switch s {
		case "minutes":
			return "fa-file-alt"
		case "bulletin":
			return "fa-newspaper"
		case "constitution":
			return "fa-scroll"
		case "form":
			return "fa-file-signature"
		case "report":
			return "fa-chart-bar"
		case "financial":
			return "fa-file-invoice-dollar"
		default:
			return "fa-file"
		}
	},
	"humanFileSize": func(size int64) string {
		const unit = 1024
		if size < unit {
			return fmt.Sprintf("%d B", size)
		}
		div, exp := int64(unit), 0
		for n := size / unit; n >= unit; n /= unit {
			div *= unit
			exp++
		}
		return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
	},
	"prayerStatusLabel": func(s string) string {
		switch s {
		case "active":
			return "Active"
		case "answered":
			return "Answered"
		case "closed":
			return "Closed"
		default:
			return s
		}
	},
	"statusOptions": func() [][2]string {
		return [][2]string{
			{"new", "New"},
			{"contacted", "Contacted"},
			{"follow_up_scheduled", "Follow-up scheduled"},
			{"follow_up_done", "Follow-up done"},
			{"converted", "Converted"},
			{"no_response", "No response"},
		}
	},
	"howHeardLabel": func(s string) string {
		switch s {
		case "walk_in":
			return "Walk-in"
		case "invited_by_member":
			return "Invited by member"
		case "social_media":
			return "Social media"
		case "website":
			return "Website"
		case "flyer":
			return "Flyer / poster"
		case "other":
			return "Other"
		default:
			return s
		}
	},
	"careTypeLabel": func(s string) string {
		switch s {
		case "visit":
			return "Pastoral Visit"
		case "counseling":
			return "Counseling"
		case "phone_call":
			return "Phone Call"
		case "prayer_session":
			return "Prayer Session"
		case "hospital_visit":
			return "Hospital Visit"
		case "bereavement":
			return "Bereavement Support"
		case "other":
			return "Other"
		default:
			return s
		}
	},
	"careTypeIcon": func(s string) string {
		switch s {
		case "visit":
			return "fa-house-user"
		case "counseling":
			return "fa-comments"
		case "phone_call":
			return "fa-phone"
		case "prayer_session":
			return "fa-hands-praying"
		case "hospital_visit":
			return "fa-hospital"
		case "bereavement":
			return "fa-heart"
		default:
			return "fa-hand-holding-heart"
		}
	},
	"careTypeOptions": func() [][2]string {
		return [][2]string{
			{"visit", "Pastoral Visit"},
			{"counseling", "Counseling"},
			{"phone_call", "Phone Call"},
			{"prayer_session", "Prayer Session"},
			{"hospital_visit", "Hospital Visit"},
			{"bereavement", "Bereavement Support"},
			{"other", "Other"},
		}
	},
	"deptTypeLabel": func(s string) string {
		switch s {
		case "worship":
			return "Worship & Music"
		case "youth":
			return "Youth Ministry"
		case "children":
			return "Children's Ministry"
		case "outreach":
			return "Outreach & Evangelism"
		case "administration":
			return "Administration"
		case "finance":
			return "Finance & Stewardship"
		case "media":
			return "Media & Communications"
		default:
			return "Other"
		}
	},
	"deptTypeIcon": func(s string) string {
		switch s {
		case "worship":
			return "fa-music"
		case "youth":
			return "fa-person-running"
		case "children":
			return "fa-child"
		case "outreach":
			return "fa-earth-africa"
		case "administration":
			return "fa-building"
		case "finance":
			return "fa-coins"
		case "media":
			return "fa-video"
		default:
			return "fa-layer-group"
		}
	},
	"deptTypeOptions": func() [][2]string {
		return [][2]string{
			{"worship", "Worship & Music"},
			{"youth", "Youth Ministry"},
			{"children", "Children's Ministry"},
			{"outreach", "Outreach & Evangelism"},
			{"administration", "Administration"},
			{"finance", "Finance & Stewardship"},
			{"media", "Media & Communications"},
			{"other", "Other"},
		}
	},
	"milestoneTypeLabel": func(s string) string {
		switch s {
		case "baby_dedication":
			return "Baby Dedication"
		case "confirmation":
			return "Confirmation"
		case "membership":
			return "Church Membership"
		case "marriage":
			return "Marriage"
		case "ordination":
			return "Ordination"
		default:
			return "Other"
		}
	},
	"milestoneTypeIcon": func(s string) string {
		switch s {
		case "baby_dedication":
			return "fa-baby"
		case "confirmation":
			return "fa-hands-praying"
		case "membership":
			return "fa-id-card"
		case "marriage":
			return "fa-rings-wedding"
		case "ordination":
			return "fa-cross"
		default:
			return "fa-star"
		}
	},
	"milestoneTypeOptions": func() [][2]string {
		return [][2]string{
			{"baby_dedication", "Baby Dedication"},
			{"confirmation", "Confirmation"},
			{"membership", "Church Membership"},
			{"marriage", "Marriage"},
			{"ordination", "Ordination"},
			{"other", "Other"},
		}
	},
	"relationTypeOptions": func() [][2]string {
		return [][2]string{
			{"wife", "Wife"}, {"husband", "Husband"},
			{"mother", "Mother"}, {"father", "Father"},
			{"daughter", "Daughter"}, {"son", "Son"},
			{"sister", "Sister"}, {"brother", "Brother"},
			{"grandmother", "Grandmother"}, {"grandfather", "Grandfather"},
			{"granddaughter", "Granddaughter"}, {"grandson", "Grandson"},
			{"aunt", "Aunt"}, {"uncle", "Uncle"},
			{"niece", "Niece"}, {"nephew", "Nephew"},
			{"cousin", "Cousin"},
			{"friend", "Friend"},
			{"godmother", "Godmother"}, {"godfather", "Godfather"},
			{"goddaughter", "Goddaughter"}, {"godson", "Godson"},
			{"other", "Other"},
		}
	},
	"statusLabel": func(s string) string {
		switch s {
		case "new":
			return "New"
		case "contacted":
			return "Contacted"
		case "follow_up_scheduled":
			return "Follow-up scheduled"
		case "follow_up_done":
			return "Follow-up done"
		case "converted":
			return "Converted"
		case "no_response":
			return "No response"
		default:
			return s
		}
	},
}
