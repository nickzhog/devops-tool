package redis

import (
	"fmt"
)

func prepareKey(id, mtype string) string {
	return fmt.Sprintf("metric:%s_%s", id, mtype)
}
