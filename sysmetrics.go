package sysmetrics

import (
	// #include <unistd.h>
	"C"
	"bufio"
	"bytes"
	"expvar"
	"fmt"
	"io/ioutil"
	"runtime"
	"strconv"
	"strings"
)

var sc_clk_tck float64 = float64(C.sysconf(C._SC_CLK_TCK))
var cputime cpuTimeStat = cpuTimeStat{-1, -1, "/proc/self/stat"}

func init() {
	expvar.Publish("runtime", expvar.Func(runtimestats))
	expvar.Publish("proc.stat", &procStat{"/proc/self/stat"})
	expvar.Publish("proc.cpu.seconds", expvar.Func(cputimestats))
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
}

func runtimestats() interface{} {
	return &runtimestate{
		NumCPU:       runtime.NumCPU(),
		NumCgoCall:   runtime.NumCgoCall(),
		NumGoroutine: runtime.NumGoroutine(),
	}
}

type procStat struct{ path string }

var stat_names []string = []string{
	"pid",         //01 -
	"",            //02 - procname
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
	"rsslim",      //25 -
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

	first := true
	fmt.Fprintf(&b, "{")
	for i, e := range fields {
		if i < len(stat_names) && stat_names[i] != "" {
			if !first {
				fmt.Fprintf(&b, ",")
			}
			fmt.Fprintf(&b, "%q:%v", stat_names[i], e)
		} else {
			// will pass here if kernel has new metrics
		}
		first = false
	}
	fmt.Fprintf(&b, "}")
	return b.String()
}
