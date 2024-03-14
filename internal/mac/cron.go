package mac

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// minute hour day of week numbers
type CronSpec struct {
	Minute, Hour int
	DaysOfWeek   []time.Weekday
}

func (c CronSpec) String() string {
	return fmt.Sprintf("%d %d %v", c.Minute, c.Hour, c.DaysOfWeek)
}

func ParseCronSpec(s string) (CronSpec, error) {
	var c CronSpec
	var dow string
	_, err := fmt.Sscanf(s, "%d %d %s", &c.Minute, &c.Hour, &dow)
	if strings.Contains(dow, ",") {
		dows := strings.Split(dow, ",")
		for _, dow := range dows {
			d, err := strconv.Atoi(dow)
			if err != nil {
				return c, err
			}
			c.DaysOfWeek = append(c.DaysOfWeek, time.Weekday(d))
		}
	}
	if strings.Contains(dow, "-") {
		dows := strings.Split(dow, "-")
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
	}
	return c, err
}
