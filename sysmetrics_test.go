package sysmetrics

import (
	"expvar"
	"fmt"
	"testing"
)

func Test_runtime_metrics(t *testing.T) {
	fmt.Printf("#----------------------------\n")
	dump("runtime.metrics")
	fmt.Printf("#----------------------------\n")
}

func Test_proc_stat(t *testing.T) {
	fmt.Printf("#----------------------------\n")
	dump("proc.stat")
	fmt.Printf("#----------------------------\n")
}

func Test_cputime(t *testing.T) {
	fmt.Printf("#----------------------------\n")
	dump("proc.cpu.seconds")

	for i := 0; i < 10000; i++ {
		ii := i
		go func() {
			x := 0
			for j := 0; j < 10000000; j++ {
				jj := j
				x = x + jj + ii
			}
			if x < 0 {
				x++
			} else {
				x--
			}
		}()

	}

	fmt.Printf("#--  --  --  --  --  --  -- --\n")
	dump("runtime")
	fmt.Printf("#--  --  --  --  --  --  -- --\n")
	dump("proc.cpu.seconds")
	fmt.Printf("#----------------------------\n")
}

func Test_runtime(t *testing.T) {
	fmt.Printf("#----------------------------\n")
	dump("runtime")
	fmt.Printf("#----------------------------\n")
	dump("proc.cpu.seconds")
	fmt.Printf("#----------------------------\n")
}

func dump(v string) {
	fmt.Printf("%s: %+v\n", v, expvar.Get(v))
}

func dumpall() {
	first := true
	expvar.Do(func(kv expvar.KeyValue) {
		if !first {
			fmt.Printf(",\n")
		}
		first = false
		fmt.Printf("%q: %s", kv.Key, kv.Value)
	})
}
