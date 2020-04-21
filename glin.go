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

	"github.com/raff/govaluate"
)

var (
	VERSION     = "0.12.0"
	SPACES      = regexp.MustCompile("\\s+")
	INVALID_POS = errors.New("invalid position")

	OK              = 0
	MATCH_FOUND     = 100
	MATCH_NOT_FOUND = 101

	SET = struct{}{}

	gitCommit, buildDate string
)

type Pos struct {
	Start, End *int
	Value      *string
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
			p.Value = &s
			return nil
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
	if p.Value != nil {
		return []string{*p.Value}
	}

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

func Unescape(s string) string {
	parts := strings.Split(s, `\`)
	if len(parts) == 1 {
		return s
	}

	u := parts[0]

	for _, p := range parts[1:] {
		if len(p) == 0 {
			u += "\\"
			continue
		}

		switch p[0] {
		case 'n':
			u += "\n" + p[1:]

		case 'r':
			u += "\n" + p[1:]

		case 't':
			u += "\t" + p[1:]

		default:
			u += p
		}
	}

	return u
}

func Print(format string, a []string) {
	printable := make([]interface{}, len(a))

	for i, v := range a {
		printable[i] = v
	}

	fmt.Printf(format, printable...)
}

func MustEvaluate(expr string) *govaluate.EvaluableExpression {
	ee, err := govaluate.NewEvaluableExpressionWithFunctions(expr, funcs)
	if err != nil {
		log.Fatalf("%q: %v", expr, err)
	}

	return ee
}

type Context struct {
	vars   map[string]interface{}
	fields []string
	lineno int
}

func (p *Context) Get(name string) (interface{}, error) {

	if strings.HasPrefix(name, "$") {
		n, err := strconv.Atoi(name[1:])
		if err != nil {
			return nil, err
		}

		if n < len(p.fields) {
			return p.fields[n], nil
		}

		return nil, fmt.Errorf("No field %q", name)
	}

	if name == "NF" {
		if p.fields == nil {
			return 0, nil
		}
		return len(p.fields) - 1, nil
	}

	if name == "NR" {
		return p.lineno, nil
	}

	if value, ok := p.vars[name]; ok {
		return value, nil
	}

	return nil, fmt.Errorf("No variable %q", name)
}

func (p *Context) Set(name string, value interface{}) error {

	if strings.HasPrefix(name, "$") {
		return fmt.Errorf("Cannot override field %q", name)
	}

	if name == "NF" || name == "NR" {
		return fmt.Errorf("Cannot override %v", name)
	}

	p.vars[name] = value
	return nil
}

func toFloat(arg interface{}) (float64, error) {
	switch v := arg.(type) {
	case string:
		f, err := strconv.ParseFloat(v, 64)
		return f, err
	case bool:
		if v {
			return 1.0, nil
		} else {
			return 0.0, nil
		}
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	default:
		return v.(float64), nil
	}
}

var funcs = map[string]govaluate.ExpressionFunction{
	"num": func(arguments ...interface{}) (interface{}, error) {
		if len(arguments) != 1 {
			return nil, fmt.Errorf("- one parameter expected, got %d", len(arguments))
		}

		return toFloat(arguments[0])
	},

	"int": func(arguments ...interface{}) (interface{}, error) {
		if len(arguments) != 1 {
			return nil, fmt.Errorf("- one parameter expected, got %d", len(arguments))
		}

		f, err := toFloat(arguments[0])
		return int(f), err
	},

	"len": func(arguments ...interface{}) (interface{}, error) {
		if len(arguments) != 1 {
			return nil, fmt.Errorf("- one parameter expected, got %d", len(arguments))
		}

		if s, ok := arguments[0].(string); ok {
			return float64(len(s)), nil
		}

		return nil, fmt.Errorf("- expected string, got %T", arguments[0])
	},

	"substr": func(arguments ...interface{}) (interface{}, error) {
		if len(arguments) < 2 || len(arguments) > 3 {
			return nil, fmt.Errorf("- expected substr(str, num [, num]), got %d", len(arguments))
		}

		s, ok := arguments[0].(string)
		if !ok {
			return nil, fmt.Errorf("- expected string, got %T", arguments[0])
		}

		f, err := toFloat(arguments[1])
		if err != nil {
			return nil, fmt.Errorf("- expected numbber, got %T", arguments[1])
		}

		start := int(f)
		slen := -1

		if len(arguments) == 3 {
			f, err := toFloat(arguments[2])
			if err != nil {
				return nil, fmt.Errorf("- expected numbber, got %T", arguments[2])
			}

			slen = int(f)
		}

		if start < 0 { // substr("string", -1) == substr("string", 5)
			start = len(s) + start
			if start < 0 {
				start = 0
			}
		}

		if start >= len(s) {
			return "", nil
		}

		s = s[start:]
		if slen >= 0 && slen < len(s) {
			return s[:slen], nil
		}

		return s, nil
	},
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
	contains := flag.String("contains", "", "output only lines that contains the pattern")
	format := flag.String("printf", "", "output is formatted according to specified format")
	matches := flag.String("matches", "", "return status code 100 if any line matches the specified pattern, 101 otherwise")
	after := flag.String("after", "", "process fields in line after specified tag (remove text before tag)")
	before := flag.String("before", "", "process fields in line before specified tag (remove text after tag)")
	afterline := flag.String("after-line", "", "process lines after lines that matches")
	beforeline := flag.String("before-line", "", "process lines before lines that matches")
	afterlinen := flag.Int("after-linen", 0, "process lines after n lines")
	printline := flag.Bool("line", false, "print line numbers")
	debug := flag.Bool("debug", false, "print debug info")
	exprbegin := flag.String("begin", "", "expression to be executed before processing lines")
	exprend := flag.String("end", "", "expression to be executed after processing lines")
	exprline := flag.String("expr", "", "expression to be executed for each line")
	exprtest := flag.String("test", "", "test expression (skip line if false)")
	uniq := flag.Bool("uniq", false, "print only unique lines")
	remove := flag.Bool("remove", false, "remove specified fields instead of selecting them")
	pexpr := flag.Bool("print-expr", false, "print result of -expr")

	flag.Parse()

	if *version {
		extra := ""
		if gitCommit != "" {
			extra = fmt.Sprintf(" (%.4v %v)", gitCommit, buildDate)
		}

		fmt.Printf("%s version %s%v\n", path.Base(os.Args[0]), VERSION, extra)
		return
	}

	pos := make([]Pos, len(flag.Args()))

	for i, arg := range flag.Args() {
		pos[i].Set(arg)
	}

	if len(*format) > 0 && !strings.HasSuffix(*format, "\n") {
		*format += "\n"
	}

	var split_re, split_pattern, match_pattern, grep_pattern *regexp.Regexp
	var expr_begin, expr_end, expr_line, expr_test *govaluate.EvaluableExpression

	status_code := OK

	if len(*matches) > 0 {
		match_pattern = regexp.MustCompile(*matches)
		status_code = MATCH_NOT_FOUND
	}

	if len(*grep) > 0 {
		if !strings.ContainsAny(*grep, "()") {
			*grep = "(" + *grep + ")"
		}
		grep_pattern = regexp.MustCompile(*grep)
	}

	if len(*re) > 0 {
		split_pattern = regexp.MustCompile(*re)
	}

	if len(*ire) > 0 {
		split_re = regexp.MustCompile(*ire)
	}

	if len(*exprbegin) > 0 {
		expr_begin = MustEvaluate(*exprbegin)
	}
	if len(*exprline) > 0 {
		expr_line = MustEvaluate(*exprline)
	}
	if len(*exprend) > 0 {
		expr_end = MustEvaluate(*exprend)
	}
	if len(*exprtest) > 0 {
		expr_test = MustEvaluate(*exprtest)
	}

	scanner := bufio.NewScanner(os.Stdin)
	len_after := len(*after)
	len_afterline := len(*afterline)
	uniques := map[string]struct{}{}

	expr_context := Context{vars: map[string]interface{}{}}

	if expr_begin != nil {
		_, err := expr_begin.Eval(&expr_context)
		if err != nil {
			log.Println("error in begin", err)
		}
		// else, should we print the result ?
	}

	for scanner.Scan() {
		if scanner.Err() != nil {
			log.Fatal(scanner.Err())
		}

		line := scanner.Text()
		expr_context.lineno += 1

		if *afterlinen >= expr_context.lineno {
			continue
		}

		if len_afterline > 0 {
			if strings.Contains(line, *afterline) {
				len_afterline = 0
			}

			continue
		}

		if len(*beforeline) > 0 && strings.Contains(line, *beforeline) {
			break
		}

		if len(*contains) > 0 && !strings.Contains(line, *contains) {
			continue
		}

		if len_after > 0 {
			i := strings.Index(line, *after)
			if i < 0 {
				continue // no match
			}

			line = line[i+len_after:]

			if len(*before) > 0 {
				i := strings.Index(line, *before)
				if i >= 0 {
					line = line[:i]
				}
			}
		}

		expr_context.fields = []string{line} // $0 is the full line

		if grep_pattern != nil {
			if matches := grep_pattern.FindStringSubmatch(line); matches != nil {
				expr_context.fields = matches
			} else {
				continue
			}
		} else if split_pattern != nil {
			if matches := split_pattern.FindStringSubmatch(line); matches != nil {
				expr_context.fields = matches
			}
		} else if split_re != nil {
			// split line according to input regular expression
			expr_context.fields = append(expr_context.fields, split_re.Split(line, -1)...)
		} else if *ifs == " " {
			// split line on spaces (compact multiple spaces)
			expr_context.fields = append(expr_context.fields, SPACES.Split(strings.TrimSpace(line), -1)...)
		} else {
			// split line according to input field separator
			expr_context.fields = append(expr_context.fields, strings.Split(line, *ifs)...)
		}

		if *debug {
			log.Printf("input fields: %q\n", expr_context.fields)
			if len(pos) > 0 {
				if *remove {
					log.Printf("output fields remove: %q\n", pos)
				} else {
					log.Printf("output fields: %q\n", pos)
				}
			}
		}

		var result []string

		// do some processing
		if len(pos) > 0 {
			if *remove {
				for _, p := range pos {
					result = append(result, Slice(expr_context.fields, p)...) // XXX: how to remove fields
				}
			} else {
				for _, p := range pos {
					result = append(result, Slice(expr_context.fields, p)...)
				}
			}
		} else {
			result = expr_context.fields[1:]
		}

		if *unquote {
			result = Unquote(result)
		}

		if *quote {
			result = Quote(result)
		}

		if *printline {
			fmt.Printf("%d: ", expr_context.lineno)
		}

		if expr_test != nil {
			res, err := expr_test.Eval(&expr_context)
			if err != nil {
				log.Println("error in expr", err)
			} else {
				switch test := res.(type) {
				case bool:
					if !test {
						continue
					}

				default:
					log.Println("boolean expected, got", test)
					continue
				}
			}
		}

		if *uniq {
			l := strings.Join(result, " ")
			if _, ok := uniques[l]; ok {
				continue
			}

			uniques[l] = SET
		}

		if expr_line != nil {
			res, err := expr_line.Eval(&expr_context)
			if err != nil {
				log.Println("error in expr", err)
			} else if *pexpr {
				fmt.Println(res)
			} else {
				for i, v := range result {
					if v == `{{expr}}` {
						result[i] = Unescape(fmt.Sprintf("%v", res))
					}
				}
			}
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

	if expr_end != nil {
		res, err := expr_end.Eval(&expr_context)
		if err != nil {
			log.Println("error in end", err)
		} else {
			fmt.Println(res)
		}
	}

	os.Exit(status_code)
}
