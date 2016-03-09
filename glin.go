package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
)

var (
	VERSION     = "0.10.0"
	SPACES      = regexp.MustCompile("\\s+")
	INVALID_POS = errors.New("invalid position")

	OK              = 0
	MATCH_FOUND     = 100
	MATCH_NOT_FOUND = 101
)

type Pos struct {
	Start, End *int
}

func (p Pos) String() (result string) {
	if p.Start != nil {
		result = strconv.Itoa(*p.Start)
	} else {
		result += "FIRST"
	}

	result += " TO "

	if p.End != nil {
		result += strconv.Itoa(*p.End)
	} else {
		result += "LAST"
	}

	return
}

func (p *Pos) Set(s string) error {
	p.Start = nil
	p.End = nil

	parts := strings.Split(s, ":")
	if len(parts) < 1 || len(parts) > 2 {
		return INVALID_POS
	}

	if len(parts[0]) > 0 {
		v, err := strconv.Atoi(parts[0])
		if err != nil {
			return err
		}

		p.Start = &v
	}

	if len(parts) == 1 {
		// not a slice
		// note: same pointer (to distinguish from *p.End == *p.Start that returns an empty slice)
		p.End = p.Start
	} else if len(parts[1]) > 0 {
		v, err := strconv.Atoi(parts[1])
		if err != nil {
			return err
		}

		p.End = &v
	}

	return nil
}

func Slice(source []string, p Pos) []string {
	var start, end int

	if p.Start == nil {
		start = 0
	} else if *p.Start >= len(source) {
		return source[0:0]
	} else if *p.Start < 0 {
		start = len(source) + *p.Start

		if start < 0 {
			start = 0
		}
	} else {
		start = *p.Start
	}

	if p.End == p.Start {
		// this should return source[start]
		end = start + 1
	} else if p.End == nil || *p.End >= len(source) {
		return source[start:]
	} else if *p.End < 0 {
		end = len(source) + *p.End
	} else {
		end = *p.End
	}

	if end < start {
		end = start
	}

	return source[start:end]
}

func Quote(a []string) []string {
	q := make([]string, len(a))
	for i, s := range a {
		q[i] = fmt.Sprintf("%q", s)
	}

	return q
}

func Unquote(a []string) []string {
	q := make([]string, len(a))
	for i, s := range a {
		q[i] = strings.Trim(s, `"'`)
	}

	return q
}

func Print(format string, a []string) {
	printable := make([]interface{}, len(a))

	for i, v := range a {
		printable[i] = v
	}

	fmt.Printf(format, printable...)
}

func main() {
	version := flag.Bool("version", false, "print version and exit")
	quote := flag.Bool("quote", false, "quote returned fields")
	unquote := flag.Bool("unquote", false, "quote returned fields")
	ifs := flag.String("ifs", " ", "input field separator")
	ire := flag.String("ifs-re", "", "input field separator (as regular expression)")
	ofs := flag.String("ofs", " ", "output field separator")
	re := flag.String("re", "", "regular expression for parsing input")
	grep := flag.String("grep", "", "output only lines that match the regular expression")
	format := flag.String("printf", "", "output is formatted according to specified format")
	matches := flag.String("matches", "", "return status code 100 if any line matches the specified pattern, 101 otherwise")
	after := flag.String("after", "", "process fields in line after specified tag")
	afterline := flag.String("after-line", "", "process lines after lines that matches")
	afterlinen := flag.Int("after-linen", 0, "process lines after n lines")
	printline := flag.Bool("line", false, "print line numbers")
	debug := flag.Bool("debug", false, "print debug info")

	flag.Parse()

	if *version {
		fmt.Printf("%s version %s\n", path.Base(os.Args[0]), VERSION)
		return
	}

	pos := make([]Pos, len(flag.Args()))

	for i, arg := range flag.Args() {
		pos[i].Set(arg)
	}

	if len(*format) > 0 && !strings.HasSuffix(*format, "\n") {
		*format += "\n"
	}

	var split_re *regexp.Regexp
	var split_pattern *regexp.Regexp
	var match_pattern *regexp.Regexp
	var grep_pattern *regexp.Regexp
	status_code := OK

	if len(*matches) > 0 {
		match_pattern = regexp.MustCompile(*matches)
		status_code = MATCH_NOT_FOUND
	}

	if len(*grep) > 0 {
		grep_pattern = regexp.MustCompile(*grep)
	}

	if len(*re) > 0 {
		split_pattern = regexp.MustCompile(*re)
	}

	if len(*ire) > 0 {
		split_re = regexp.MustCompile(*ire)
	}

	scanner := bufio.NewScanner(os.Stdin)
	len_after := len(*after)
	len_afterline := len(*afterline)
	lineno := 0

	for scanner.Scan() {
		if scanner.Err() != nil {
			log.Fatal(scanner.Err())
		}

		line := scanner.Text()

		lineno += 1

		if *afterlinen >= lineno {
			continue
		}

		if len_afterline > 0 {
			if strings.Contains(line, *afterline) {
				len_afterline = 0
			}

			continue
		}

		if len_after > 0 {
			i := strings.Index(line, *after)
			if i < 0 {
				continue // no match
			}

			line = line[i+len_after:]
		}

		fields := []string{line} // $0 is the full line

		if grep_pattern != nil {
			if matches := grep_pattern.FindStringSubmatch(line); matches != nil {
				fields = matches
			} else {
				continue
			}
		} else if split_pattern != nil {
			if matches := split_pattern.FindStringSubmatch(line); matches != nil {
				fields = matches
			}
		} else if split_re != nil {
			// split line according to input regular expression
			fields = append(fields, split_re.Split(line, -1)...)
		} else if *ifs == " " {
			// split line on spaces (compact multiple spaces)
			fields = append(fields, SPACES.Split(strings.TrimSpace(line), -1)...)
		} else {
			// split line according to input field separator
			fields = append(fields, strings.Split(line, *ifs)...)
		}

		if *debug {
			log.Printf("input fields: %q\n", fields)
			if len(pos) > 0 {
				log.Printf("output fields: %q\n", pos)
			}
		}

		var result []string

		// do some processing
		if len(pos) > 0 {
			result = make([]string, 0)

			for _, p := range pos {
				result = append(result, Slice(fields, p)...)
			}
		} else {
			result = fields[1:]
		}

		if *unquote {
			result = Unquote(result)
		}

		if *quote {
			result = Quote(result)
		}

		if *printline {
			fmt.Printf("%d: ", lineno)
		}

		if len(*format) > 0 {
			Print(*format, result)
		} else {
			// join the result according to output field separator
			fmt.Println(strings.Join(result, *ofs))
		}

		if match_pattern != nil && match_pattern.MatchString(line) {
			status_code = MATCH_FOUND
		}
	}

	os.Exit(status_code)
}
