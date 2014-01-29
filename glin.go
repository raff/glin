package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var (
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

func main() {
	ifs := flag.String("ifs", " ", "input field separator")
	ofs := flag.String("ofs", " ", "input field separator")
	quote := flag.Bool("quote", false, "quote returned fields")
	flag.Parse()

	pos := make([]Pos, len(flag.Args()))

	for i, arg := range flag.Args() {
		pos[i].Set(arg)
	}

	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		if scanner.Err() != nil {
			log.Fatal(scanner.Err())
		}

		line := scanner.Text()

		var fields, result []string

		// split the line according to input field separator
		if *ifs == " " {
			fields = SPACES.Split(strings.TrimSpace(line), -1)
		} else {
			fields = strings.Split(line, *ifs)
		}

		// do some processing
		if len(pos) > 0 {
			result = make([]string, 0)

			for _, p := range pos {
				val := strings.Join(Slice(fields, p), *ifs)
				result = append(result, val)
			}
		} else {
			result = fields
		}

		if *quote {
			result = Quote(result)
		}

		// join the result according to output field separator
		fmt.Println(strings.Join(result, *ofs))
	}
}
