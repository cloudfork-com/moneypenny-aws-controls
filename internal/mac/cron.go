package mac

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"
)

const daySeparator = "/"
const rangeSeparator = "-"

// minute hour day of week numbers
type CronSpec struct {
	Minute, Hour int
	DaysOfWeek   []time.Weekday
}

func (c CronSpec) String() string {
	return fmt.Sprintf("%d %d %v", c.Minute, c.Hour, c.DaysOfWeek)
}

func (c CronSpec) IsEffectiveOnWeekday(w time.Weekday) bool {
	for _, dow := range c.DaysOfWeek {
		if dow == w {
			return true
		}
	}
	return false
}

func ParseCronSpec(s string) (CronSpec, error) {
	var c CronSpec
	var dow string
	_, err := fmt.Sscanf(s, "%d %d %s", &c.Minute, &c.Hour, &dow)
	if strings.Contains(dow, daySeparator) {
		dows := strings.Split(dow, daySeparator)
		for _, dow := range dows {
			d, err := strconv.Atoi(dow)
			if err != nil {
				return c, err
			}
			wd := time.Weekday(d)
			if slices.Contains(c.DaysOfWeek, wd) {
				return c, fmt.Errorf("duplicate day of week %d", wd)
			}
			c.DaysOfWeek = append(c.DaysOfWeek, wd)
		}
	} else if strings.Contains(dow, rangeSeparator) {
		dows := strings.Split(dow, rangeSeparator)
		d1, err := strconv.Atoi(dows[0])
		if err != nil {
			return c, err
		}
		d2, err := strconv.Atoi(dows[1])
		if err != nil {
			return c, err
		}
		for d := d1; d <= d2; d++ {
			c.DaysOfWeek = append(c.DaysOfWeek, time.Weekday(d))
		}
	} else {
		// one
		d, err := strconv.Atoi(dow)
		if err != nil {
			return c, err
		}
		c.DaysOfWeek = append(c.DaysOfWeek, time.Weekday(d))
	}
	return c, err
}
