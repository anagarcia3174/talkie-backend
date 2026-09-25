package moderation

import (
	"strings"
	"unicode"
)

var blockedWords = toSet(
	// racial / ethnic slurs
	"nigger", "niggers", "nigga", "niggas", "chink", "chinks", "gook", "gooks",
	"spic", "spics", "wetback", "wetbacks", "kike", "kikes", "coon", "coons",
	"raghead", "ragheads", "towelhead", "towelheads", "beaner", "beaners",
	"paki", "pakis",
	"niggah", "nigguh", "niggaz", "niglet", "niglets", "jigaboo", "jigaboos",
	"porchmonkey", "junglebunny", "zipperhead", "zipperheads", "chingchong",
	"kyke", "kykes", "hymie", "hymies",
	// homophobic / transphobic slurs
	"faggot", "faggots", "fag", "fags", "dyke", "dykes", "tranny", "trannies",
	"fudgepacker", "fudgepackers", "faggy", "shemale", "shemales",
	// ableist slurs
	"retard", "retards", "retarded", "mongoloid", "mongoloids",
	"kys",
)

var blockedPhrases = []string{
	"porch monkey",
	"porch monkeys",
	"jungle bunny",
	"jungle bunnies",
	"ching chong",
	"kill yourself",
	"kill yourselves",
	"kill urself",
	"kill your self",
	"neck yourself",
	"unalive yourself",
	"hope you die",
	"drink bleach",
}

var leet = map[rune]rune{
	'0': 'o', '1': 'i', '3': 'e', '4': 'a', '@': 'a',
	'5': 's', '$': 's', '7': 't',
}

func ContainsObjectionable(text string) bool {
	words := normalize(text)
	for _, w := range words {
		for _, v := range variants(w) {
			if _, ok := blockedWords[v]; ok {
				return true
			}
		}
	}

	joined := " " + strings.Join(words, " ") + " "
	for _, p := range blockedPhrases {
		if strings.Contains(joined, " "+p+" ") {
			return true
		}
	}
	return false
}

func normalize(text string) []string {
	var b strings.Builder
	for _, r := range strings.ToLower(text) {
		if m, ok := leet[r]; ok {
			r = m
		}
		if unicode.IsLetter(r) {
			b.WriteRune(r)
		} else {
			b.WriteRune(' ')
		}
	}

	raw := strings.Fields(b.String())
	words := make([]string, 0, len(raw))
	var singles strings.Builder
	flush := func() {
		if singles.Len() > 0 {
			words = append(words, singles.String())
			singles.Reset()
		}
	}
	for _, w := range raw {
		if len([]rune(w)) == 1 {
			singles.WriteString(w)
			continue
		}
		flush()
		words = append(words, w)
	}
	flush()
	return words
}

func variants(w string) []string {
	return []string{w, collapseRuns(w, 2), collapseRuns(w, 1)}
}

func collapseRuns(w string, keep int) string {
	rs := []rune(w)
	var b strings.Builder
	for i := 0; i < len(rs); {
		j := i
		for j < len(rs) && rs[j] == rs[i] {
			j++
		}
		n := j - i
		if n >= 3 {
			n = keep
		}
		for k := 0; k < n; k++ {
			b.WriteRune(rs[i])
		}
		i = j
	}
	return b.String()
}

func toSet(words ...string) map[string]struct{} {
	m := make(map[string]struct{}, len(words))
	for _, w := range words {
		m[w] = struct{}{}
	}
	return m
}
