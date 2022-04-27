package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"strconv"
)

var (
	flagSyscalls = flag.String("e", "", "only trace specified syscalls")
	flagOutput   = flag.String("o", "stracefile.json", "json output file")
	flagTimeout  = flag.Int64("t", 10, "strace timeout (secs)")
)

var (
	// -f trace child processes
	// -T time spent in each syscall
	// -ttt timestamp of each event (microseconds)
	// -qq don't display process exit status
	defaultStraceArgs = []string{"-f", "-T", "-ttt", "-qq"}
)

func main() {

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] command\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
	}

	flag.Parse()

	if len(flag.Args()) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	// run strace
	userStraceArgs := []string{}
	if *flagSyscalls != "" {
		userStraceArgs = append(userStraceArgs, "-e", *flagSyscalls)
	}
	userStraceArgs = append(userStraceArgs, flag.Args()...)

	tmp, err := os.CreateTemp("", "stracefile")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(tmp.Name())

	defaultStraceArgs = append(defaultStraceArgs, "-o", tmp.Name())

	strace := Strace{
		DefaultArgs: defaultStraceArgs,
		UserArgs:    userStraceArgs,
		Timeout:     *flagTimeout,
	}
	strace.Run()

	// parse results
	var events []*Event
	preserved := make(map[string]*Event) // [pid+syscall]*Event
	scanner := bufio.NewScanner(tmp)

	for scanner.Scan() {
		e := NewEvent(scanner.Text())
		switch {
		case e.Cat == "unfinished":
			k := strconv.Itoa(e.Pid) + e.Name
			preserved[k] = e
			break
		case e.Cat == "detached":
			k := strconv.Itoa(e.Pid) + e.Name
			p := preserved[k]
			e.Args.First = p.Args.First
			events = append(events, e)
			delete(preserved, k)
			break
		case e.Cat == "other":
			break
		default:
			events = append(events, e)
		}
	}
	// add any unfinished/preserved traces to events
	for _, p := range preserved {
		p.Ph = "i" // instant event
		events = append(events, p)
	}

	// save results
	te := TraceEvents{events}
	te.Save(*flagOutput)

	fmt.Printf("[+] Trace file saved to: %s\n", *flagOutput)
	fmt.Printf("[+] Analyze results: %s\n", "https://ui.perfetto.dev/")

}
