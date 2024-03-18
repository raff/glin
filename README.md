glin
====

Go line scanner

Glin is similar in concept to "awk" but offers a more limited and simplified interface.

Glin reads lines from standard input, splits each line into fields, and writes a re-ordered list of fields.

A common usage for glin is to "filter" a log file and return only a limited set of fields.

usage
=====

    glin  [options] [output field list]
    
    Options:
      -after string
            process fields in line after specified tag (remove text before tag)
      -after-line string
            process lines after lines that matches
      -after-linen int
            process lines after n lines
      -before string
            process fields in line before specified tag (remove text after tag)
      -before-line string
            process lines before lines that matches
      -begin string
            expression to be executed before processing lines
      -contains string
            output only lines that contains the pattern
      -debug
            print debug info
      -end string
            expression to be executed after processing lines
      -expr string
            expression to be executed for each line
      -grep string
            output only lines that match the regular expression
      -ifn int
            maximum number of fields when splitting input (<1=all) (default -1)
      -ifs string
            input field separator (default " ")
      -ifs-re string
            input field separator (as regular expression)
      -line
            print line numbers
      -matches string
            return status code 100 if any line matches the specified pattern, 101 otherwise
      -ofs string
            output field separator (default " ")
      -print-expr
            print result of -expr
      -printf string
            output is formatted according to specified format
      -quote
            quote returned fields
      -re string
            regular expression for parsing input
      -remove
            remove specified fields instead of selecting them
      -shlex
            split using shlex/shell-style rules
      -test string
            test expression (skip line if false)
      -ucount
            print unique lines (and count)
      -uniq
            print only unique lines
      -unquote
            quote returned fields
      -version
            print version and exit

    Output field list: one or more indices (or slices) of fields to return. Also, use `{{expr}}` to return the current value of `-expr` or `{{expr:varname}}` to return the value of variable `varname`.
    
Expressions use the syntax described [here](https://github.com/Knetic/govaluate/blob/master/MANUAL.md) with the following additions:
* It's possible to set variables:

      name=expr

  And use variables:

      $name

* The following variables are predefined (same as `awk`):
  - $NR : current number of records / lines
  - $NF : number of fields in the current line
  - $0 : the current line
  - $1 to $NF: the value of the numbered field

* The following functions are available:
  - print(args...) : print one or more arguments
  - format(fmt, args...) : format one or more arguments according to `fmt` (like `sprintf`)
  - num("string") : convert `string` to float, to apply numeric operations
  - int("string") : convert `string` to int, to apply numberic operations
  - len("string") : return length of `string`
  - substr("string", start, [len]) : return substring of `string` starting from `start` (truncate to `len` characters if `len is specified)

Also:
* Slices follow the Go slice convention (start:end) or better, the Python slice convention (negative values are offsets from the end, so -1 is the last field).
* Field 0 is the input line (as in "awk") so the various fields are indexed from 1 to the number of fields.
