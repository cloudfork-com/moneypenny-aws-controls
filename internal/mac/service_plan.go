package mac

import (
	"errors"
	"fmt"
	"strings"
)

type ServicePlan struct {
	Service
	TagValue     string         `json:"moneypenny-tag-value"`
	StateChanges []*StateChange `json:"state-changes"`
}
type StateChange struct {
	DesiredState string `json:"desired-state"`
	Cron         string `json:"cron"`
	CronSpec     CronSpec
}

func (s StateChange) String() string {
	return fmt.Sprintf("%s=%s.", s.DesiredState, s.Cron)
}

func (t *ServicePlan) Validate() error {
	if t.TagValue != "" {
		chgs, err := ParseStateChanges(t.TagValue)
		if err != nil {
			return err
		}
		t.StateChanges = chgs
		return nil
	}
	for _, each := range t.StateChanges {
		spec, err := ParseCronSpec(each.Cron)
		if err != nil {
			return err
		}
		each.CronSpec = spec
	}
	return nil
}

// running=0 8 1-5. stopped=0 18 1-5.
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
		switch stateParts[0] {
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
		default:
			return list, errors.New("unknown state:" + stateParts[0])
		}
	}
	return
}
