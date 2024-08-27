//go:build ignore

package sysmetrics

import (
	// #include <unistd.h>
	"C"
)

var sc_clk_tck float64 = float64(C.sysconf(C._SC_CLK_TCK))
