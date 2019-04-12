//go:generate propgen
package main

import "time"

type Body struct {
	pos, vec complex128
	t        time.Time
	Public   int
}
