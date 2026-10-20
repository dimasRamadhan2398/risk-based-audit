package activityplanform

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

// List limits for Justifikasi, Tujuan and Fokus (contract 6.4).
const (
	MaxListItems   = 20
	MaxListItemLen = 1000
)

// leadingNumber is a list marker such as "1." or "2)" at the start of an item.
var leadingNumber = regexp.MustCompile(`^\s*\d+[.)]\s*`)

// Normalize trims every item, strips a leading "1." / "1)" marker and drops
// empty items. A multi-line item is split into one item per line first, so a
// pasted numbered list becomes separate items. The frontend uses the same rules
// (utils/numberedList.ts).
func Normalize(items []string) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		for _, line := range strings.Split(strings.ReplaceAll(item, "\r\n", "\n"), "\n") {
			line = strings.TrimSpace(leadingNumber.ReplaceAllString(line, ""))
			if line != "" {
				out = append(out, line)
			}
		}
	}
	return out
}

// Join stores normalised items as one text value, one item per line.
func Join(items []string) string {
	return strings.Join(items, "\n")
}

// Split reads a stored text value (or a legacy PKAT free-text field) back into
// normalised items.
func Split(text string) []string {
	return Normalize([]string{text})
}

// checkList validates normalised items: at least 1, at most MaxListItems, each
// at most MaxListItemLen characters. field names the request field.
func checkList(row int, field, label string, items []string) error {
	if len(items) == 0 {
		return invalidRow(row, field, fmt.Sprintf("assignment %d: %s needs at least one item", row, label))
	}
	if len(items) > MaxListItems {
		return invalidRow(row, field, fmt.Sprintf("assignment %d: %s allows at most %d items", row, label, MaxListItems))
	}
	for _, it := range items {
		if utf8.RuneCountInString(it) > MaxListItemLen {
			return invalidRow(row, field, fmt.Sprintf("assignment %d: %s items are at most %d characters", row, label, MaxListItemLen))
		}
	}
	return nil
}
