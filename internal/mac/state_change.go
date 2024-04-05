package mac

import (
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
)

type StateChange struct {
	DesiredState string `json:"desired-state"`
	DesiredCount int    `json:"desired-count"`
	Cron         string `json:"cron"`
	CronSpec     CronSpec
}

func (s *StateChange) String() string {
	return fmt.Sprintf("%s=%s.", s.DesiredState, s.Cron)
}

// running=0 8 1-5. stopped=0 18 1-5. count=2.
func ParseStateChanges(input string) (list []*StateChange, err error) {
	changeParts := strings.Split(strings.TrimSpace(input), ".")
	for _, each := range changeParts {
		if len(each) == 0 {
			continue
		}
		stateParts := strings.Split(strings.TrimSpace(each), "=")
		if len(stateParts) != 2 {
			return list, fmt.Errorf("expected: state=expression. parts:%v", stateParts)
		}
		stateName := stateParts[0]
		if strings.HasPrefix(stateName, "//") {
			break
		}
		switch stateName {
		case "running":
			expr := strings.Trim(stateParts[1], ".")
			spec, err := ParseCronSpec(expr)
			if err != nil {
				return list, fmt.Errorf("invalid spec for running:%w, expression:%q", err, expr)
			}
			list = append(list, &StateChange{
				DesiredState: Running,
				Cron:         expr,
				CronSpec:     spec,
			})
		case "stopped":
			expr := strings.Trim(stateParts[1], ".")
			spec, err := ParseCronSpec(expr)
			if err != nil {
				return list, fmt.Errorf("invalid spec for stopped:%w, expression:%q", err, expr)
			}
			list = append(list, &StateChange{
				DesiredState: Stopped,
				Cron:         expr,
				CronSpec:     spec,
			})
		case "count":
			expr := strings.Trim(stateParts[1], ".")
			c, err := strconv.Atoi(expr)
			if err != nil {
				return list, fmt.Errorf("invalid spec for count:%w, expression:%q", err, expr)
			}
			// find running change to update its count
			var run *StateChange
			for _, each := range list {
				if each.DesiredState == Running {
					run = each
					break
				}
			}
			if run == nil {
				slog.Warn("no running change specified for count", "moneypenny", input)
			} else {
				run.DesiredCount = c
			}
		default:
			return list, errors.New("unknown state:" + stateParts[0])
		}
	}
	return
}
