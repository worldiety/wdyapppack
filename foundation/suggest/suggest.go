// Package suggest finds the name somebody probably meant.
//
// It is shared by the command line layer, which answers a mistyped command or
// option, and by the contexts, which answer a mistyped category or
// configuration field. It sits apart from both so that neither has to import
// the other.
package suggest

import "strings"

// Closest returns the candidate closest to what was typed, if any is close
// enough to be a plausible typo.
func Closest(typed string, candidates []string) (string, bool) {
	typed = strings.ToLower(typed)
	best, bestDist := "", -1
	for _, c := range candidates {
		lc := strings.ToLower(c)
		d := distance(typed, lc)
		if strings.HasPrefix(lc, typed) && len(typed) >= 3 {
			d = min(d, 1)
		}
		if bestDist < 0 || d < bestDist {
			best, bestDist = c, d
		}
	}
	limit := max(2, len(typed)/3)
	if bestDist < 0 || bestDist > limit || bestDist >= len(typed) {
		return "", false
	}
	return best, true
}

// distance is the Levenshtein distance between two strings.
func distance(a, b string) int {
	ra, rb := []rune(a), []rune(b)
	prev := make([]int, len(rb)+1)
	cur := make([]int, len(rb)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(ra); i++ {
		cur[0] = i
		for j := 1; j <= len(rb); j++ {
			cost := 1
			if ra[i-1] == rb[j-1] {
				cost = 0
			}
			cur[j] = min(prev[j]+1, cur[j-1]+1, prev[j-1]+cost)
		}
		prev, cur = cur, prev
	}
	return prev[len(rb)]
}
