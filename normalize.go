package go_routeros

import (
	"strings"
)

func NormalizeToCommandLine(menu string, sentence ...string) string {
	if len(sentence) == 0 {
		return menu
	}

	type group struct {
		type_ string
		conds []string
	}

	var groups []group
	var current group
	current.type_ = "and"
	var preArgs []string

	for _, s := range sentence {
		switch s {
		case "?#|":
			current.type_ = "or"
			if len(current.conds) > 0 {
				groups = append(groups, current)
			}
			current = group{type_: "and"}
		case "?#&":
			if len(current.conds) > 0 {
				groups = append(groups, current)
			}
			current = group{type_: "and"}
		case "?#()":
			continue
		default:
			if strings.HasPrefix(s, "=") {
				preArgs = append(preArgs, strings.TrimPrefix(s, "="))
			} else {
				s = strings.TrimPrefix(s, "?")
				current.conds = append(current.conds, s)
			}
		}
	}

	if len(current.conds) > 0 {
		groups = append(groups, current)
	}

	var parts []string
	for _, g := range groups {
		if len(g.conds) == 0 {
			continue
		}
		groupExpr := strings.Join(g.conds, " "+g.type_+" ")
		// só coloca parênteses se houver múltiplos grupos
		if len(groups) > 1 && len(g.conds) > 1 {
			parts = append(parts, "("+groupExpr+")")
		} else {
			parts = append(parts, groupExpr)
		}
	}

	if len(parts) == 0 {
		return menu
	}

	segments := strings.SplitAfter(menu, "/")
	cmd := strings.TrimSpace(segments[len(segments)-1])

	pre := strings.TrimRight(" "+strings.Join(preArgs, " "), " ")
	switch cmd {
	case "print":
		return menu + " where " + strings.Join(parts, " and ")
	case "set", "remove", "unset", "disable", "enable":
		return menu + pre + " [find " + strings.Join(parts, " and ") + "]"
	default:
		return menu + " " + strings.Join(parts, " ")
	}
}
