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
	VERSION     = "0.9.0"
	SPACES      = regexp.MustCompile("\\s+")
	INVALID_POS = errors.New("invalid position")
)

type Pos struct {
	Start, End *int
}

func (p Pos) String() (result string) {
	if p.Start != nil {
		result = strconv.Itoa(*p.Start)
	}

	result += ":"

	if p.End != nil {
		result += strconv.Itoa(*p.End)
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
	ofs := flag.String("ofs", " ", "input field separator")
	format := flag.String("printf", "", "output is formatted according to specified format")

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

	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		if scanner.Err() != nil {
			log.Fatal(scanner.Err())
		}

		line := scanner.Text()
		fields := []string{line} // $0 is the full line

		// split the line according to input field separator
		if *ifs == " " {
			fields = append(fields, SPACES.Split(strings.TrimSpace(line), -1)...)
		} else {
			fields = append(fields, strings.Split(line, *ifs)...)
		}

		var result []string

		// do some processing
		if len(pos) > 0 {
			result = make([]string, 0)

			for _, p := range pos {
				val := strings.Join(Slice(fields, p), *ifs)
				result = append(result, val)
			}
		} else {
			result = fields[0:1]
		}

		if *unquote {
			result = Unquote(result)
		}

		if *quote {
			result = Quote(result)
		}

		if len(*format) > 0 {
			Print(*format, result)
		} else {
			// join the result according to output field separator
			fmt.Println(strings.Join(result, *ofs))
		}
	}
}
