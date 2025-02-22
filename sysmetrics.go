package sysmetrics

import (
	"bufio"
	"bytes"
	"expvar"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"runtime/metrics"
	"strconv"
	"strings"
	"sync"
)

const (
	// CLK_TCK is a constant on Linux for all architectures except alpha and ia64.
	// See e.g.
	// https://git.musl-libc.org/cgit/musl/tree/src/conf/sysconf.c#n30
	// https://github.com/containerd/cgroups/pull/12
	// https://lore.kernel.org/lkml/agtlq6$iht$1@penguin.transmeta.com/
	_SYSTEM_CLK_TCK = 100
)

var sc_clk_tck float64 = float64(_SYSTEM_CLK_TCK)
var cputime cpuTimeStat = cpuTimeStat{-1, -1, "/proc/self/stat"}

func init() {
	expvar.Publish("runtime", expvar.Func(runtimestats))
	expvar.Publish("proc.stat", &procStat{"/proc/self/stat"})
	expvar.Publish("proc.cpu.seconds", expvar.Func(cputimestats))
	expvar.Publish("runtime.metrics", New_runtimeMetrics())
}

type cpuTimeStat struct {
	utime, stime int64
	path         string
}

type cputimeT struct {
	User   float64 `json:"user"`
	System float64 `json:"system"`
	Total  float64 `json:"total"`
}

func cputimestats() interface{} {
	//var b bytes.Buffer
	contents, err := ioutil.ReadFile(cputime.path)
	if err != nil {
		return &struct{}{}
	}

	reader := bufio.NewReader(bytes.NewBuffer(contents))
	line, _, err := reader.ReadLine()
	fields := strings.Fields(string(line))
	if len(fields) < 15 {
		return &struct{}{}
	}

	utime, _ := strconv.ParseFloat(fields[13], 0)
	stime, _ := strconv.ParseFloat(fields[14], 0)

	return cputimeT{
		utime / sc_clk_tck,
		stime / sc_clk_tck,
		(utime + stime) / sc_clk_tck}
}

type runtimestate struct {
	NumCPU       int
	NumCgoCall   int64
	NumGoroutine int
	Version      string
}

func runtimestats() interface{} {
	return &runtimestate{
		NumCPU:       runtime.NumCPU(),
		NumCgoCall:   runtime.NumCgoCall(),
		NumGoroutine: runtime.NumGoroutine(),
		Version:      runtime.Version(),
	}
}

type procStat struct{ path string }

var stat_names []string = []string{
	"pid",         //01 -
	"",            //02 - comm (truncated to 16 bytes, in ())
	"",            //03 - state
	"ppid",        //04 -
	"pgrp",        //05 -
	"session",     //06 -
	"tty_nr",      //07 -
	"tpgid",       //08 -
	"flags",       //09 -
	"minflt",      //10 -
	"cminflt",     //11 -
	"majflt",      //12 -
	"cmajflt",     //13 -
	"utime",       //14 -
	"stime",       //15 -
	"cutime",      //16 -
	"cstime",      //17 -
	"priority",    //18 -
	"nice",        //19 -
	"num_threads", //20 -
	"itrealvalue", //21 -
	"starttime",   //22 -
	"vsize",       //23 -
	"rss",         //24 -
	"",            //25 - rsslim
	"",            //26 - startcode
	"",            //27 - endcode
	"",            //28 - startstack
	"",            //29 - kstkesp
	"",            //30 - kstkeip
	"",            //31 - signal
	"",            //32 - blocked
	"",            //33 - sigignore
	"",            //34 - sigcatch
	"",            //35 - wchan
	"",            //36 - nswap
	"",            //37 - cnswap
	"",            //38 - exit_signal
	"",            //39 - processor
	"",            //40 - rt_priority
	"",            //41 - policy
	"",            //42 - delayacct_blkio_ticks
	"guest_time",  //43 - guest_time
	"cguest_time", //44 - cguest_time
	"",            //45 - start_data
	"",            //46 - end_data
	"",            //47 - start_brk
	"",            //48 - arg_start
	"",            //49 - arg_end
	"",            //50 - env_start
	"",            //51 - env_end
	"",            //52 - exit_code
}

func (v *procStat) String() string {
	var b bytes.Buffer
	contents, err := ioutil.ReadFile(v.path)
	if err != nil {
		return "{}"
	}

	reader := bufio.NewReader(bytes.NewBuffer(contents))
	line, _, err := reader.ReadLine()
	fields := strings.Fields(string(line))

	sep := ""
	fmt.Fprintf(&b, "{")
	for i, e := range fields {
		if i < len(stat_names) && stat_names[i] != "" {
			fmt.Fprintf(&b, "%s%q:%v", sep, stat_names[i], e)
			sep = ",\n"
			if stat_names[i] == "rss" {
				rss, err := strconv.Atoi(e)
				if err != nil {
					continue
				}
				fmt.Fprintf(&b, `%s"rssBytes":%v`, sep, rss*os.Getpagesize())
				sep = ",\n"
			}
		} else {
			// will pass here if kernel has new metrics
		}
	}
	fmt.Fprintf(&b, "}")
	return b.String()
}

type runtimeMetrics struct {
	Samples []metrics.Sample
	sync.Mutex
}

func New_runtimeMetrics() *runtimeMetrics {
	descs := metrics.All()
	v := &runtimeMetrics{}
	samples := make([]metrics.Sample, len(descs))
	for i := range samples {
		samples[i].Name = descs[i].Name
	}
	v.Samples = samples
	return v
}

func (v *runtimeMetrics) String() string {

	var b bytes.Buffer

	v.Lock()
	defer v.Unlock()
	metrics.Read(v.Samples)

	sep := ""
	fmt.Fprintf(&b, "{")
	for _, sample := range v.Samples {
		name, value := sample.Name, sample.Value
		switch value.Kind() {
		case metrics.KindUint64:
			fmt.Fprintf(&b, "%s%q: %d", sep, name, value.Uint64())
			sep = ",\n"
		case metrics.KindFloat64:
			fmt.Fprintf(&b, "%s%q: %g", sep, name, value.Float64())
			sep = ",\n"
		case metrics.KindFloat64Histogram:
		case metrics.KindBad:
		default:
		}
	}

	fmt.Fprintf(&b, "}")
	return b.String()
}
