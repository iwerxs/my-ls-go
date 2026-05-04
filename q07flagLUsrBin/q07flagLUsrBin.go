// q07flagLUsrBin.go
package main

import (
	"fmt"
	"os"
	"os/user"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"golang.org/x/text/collate"
	"golang.org/x/text/language"
)

type fileRow struct {
	mode       string
	links      string
	owner      string
	group      string
	size       string
	date       string
	name       string
	linkTarget string
}

const halfAverageYear = 31556952 * time.Second / 2

func main() {
	target := "/usr/bin"
	if len(os.Args) > 1 && os.Args[len(os.Args)-1][0] != '-' {
		target = os.Args[len(os.Args)-1]
	}

	entries, err := os.ReadDir(target)
	if err != nil {
		return
	}

	now := time.Now()
	rows := make([]fileRow, 0, len(entries))
	var totalBlocks int64
	maxL, maxU, maxG, maxS := 0, 0, 0, 0

	for _, e := range entries {
		name := e.Name()
		// Skip hidden entries
		if len(name) > 0 && name[0] == '.' {
			continue
		}

		path := target
		if !strings.HasSuffix(path, "/") {
			path += "/"
		}
		path += name

		info, err := os.Lstat(path)
		if err != nil {
			continue
		}

		stat, ok := info.Sys().(*syscall.Stat_t)
		if !ok {
			continue
		}

		// Standard ls -l block rounding
		totalBlocks += (stat.Blocks + 0) / 2

		u, _ := user.LookupId(strconv.Itoa(int(stat.Uid)))
		g, _ := user.LookupGroupId(strconv.Itoa(int(stat.Gid)))

		r := fileRow{
			mode:  formatMode(info),
			links: strconv.FormatUint(uint64(stat.Nlink), 10),
			owner: u.Username,
			group: g.Name,
			size:  strconv.FormatInt(info.Size(), 10),
			date:  formatTime(info.ModTime(), now),
			name:  name,
		}

		if info.Mode()&os.ModeSymlink != 0 {
			if t, err := os.Readlink(path); err == nil {
				r.linkTarget = " -> " + t
			}
		}

		// Calculate column widths
		if len(r.links) > maxL {
			maxL = len(r.links)
		}
		if len(r.owner) > maxU {
			maxU = len(r.owner)
		}
		if len(r.group) > maxG {
			maxG = len(r.group)
		}
		if len(r.size) > maxS {
			maxS = len(r.size)
		}

		rows = append(rows, r)
	}

	// Locale-aware sorting (matches ls behavior)
	collator := collate.New(language.English, collate.Loose)
	sort.Slice(rows, func(i, j int) bool {
		return collator.CompareString(rows[i].name, rows[j].name) < 0
	})

	// Final Output
	fmt.Printf("total %d\n", totalBlocks)
	for _, r := range rows {
		// EXACT PRINTF: 8 verbs, 10 arguments.
		// One space between date (%s) and name (%s).
		fmt.Printf("%s %*s %s %s %*s %s %s%s\n",
			r.mode,
			maxL, r.links,
			r.owner,
			r.group,
			maxS, r.size,
			r.date,
			r.name,
			r.linkTarget,
		)
	}
}

func formatMode(fi os.FileInfo) string {
	var m string
	switch {
	case fi.IsDir():
		m = "d"
	case fi.Mode()&os.ModeSymlink != 0:
		m = "l"
	default:
		m = "-"
	}

	mode := fi.Mode()
	m += permissionTriplet(mode, 0400, 0200, 0100, os.ModeSetuid, "s", "S")
	m += permissionTriplet(mode, 0040, 0020, 0010, os.ModeSetgid, "s", "S")
	m += permissionTriplet(mode, 0004, 0002, 0001, os.ModeSticky, "t", "T")
	return m + "."
}

func permissionTriplet(mode, read, write, execute, special os.FileMode, specialSet, specialUnset string) string {
	text := ""
	if mode&read != 0 {
		text += "r"
	} else {
		text += "-"
	}
	if mode&write != 0 {
		text += "w"
	} else {
		text += "-"
	}
	if mode&special != 0 {
		if mode&execute != 0 {
			text += specialSet
		} else {
			text += specialUnset
		}
	} else if mode&execute != 0 {
		text += "x"
	} else {
		text += "-"
	}
	return text
}

func formatTime(t time.Time, now time.Time) string {
	if now.Sub(t) > halfAverageYear || t.After(now.Add(time.Hour)) {
		// Exactly two standard spaces before the year
		return t.Format("Jan _2  2006")
	}
	return t.Format("Jan _2 15:04")
}
