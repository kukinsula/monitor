package metric

import (
	"fmt"
	"math"
)

type Metric interface {
	Update() error
	Public() map[string]interface{}
	String() string
}

func checkSscanf(field string, err error, n, expected int) error {
	if err != nil {
		return fmt.Errorf("Sscanf '%s' failed: %s", field, err)
	}

	if n != expected {
		return fmt.Errorf("Sscanf '%s' parsed %d item(s) but expected %d",
			field, n, expected)
	}

	return nil
}

type kbyte int

func (k kbyte) String() string {
	var str string
	fKbyes := float64(k)

	if math.Abs(fKbyes) < 100000 {
		str = fmt.Sprintf("%d kB", int(k))
	} else if math.Abs(fKbyes) < 100000000 {
		str = fmt.Sprintf("%.3f MB", fKbyes/float64(1000))
	} else {
		str = fmt.Sprintf("%.3f GB", fKbyes/float64(1000000))
	}

	return str
}
