glin
====

Go line scanner

Glin is similar in concept to "awk" but offer a more limited and simplified interface.

Glin reads lines from stanrd input, splits each line into fields, and writes a re-ordered list of fields.

A common usage for glin is to "filter" a log file and return only a limited set of fields.

usage
=====

    glin  [--ifs={input field separator}] [--after={pattern}] [--re={pattern}] [--matches={pattern}] \
       [--ofs={output field separator}] [--quote]  [--printf={format-string] [output field list]
    
where:
* --ifs=c: the character (or string) c is used as input field separator (default space)
* --after=pattern: matches on lines containing pattern and process only the part of the line after pattern
* --re=pattern: split lines according to "pattern" and return the result of pattern.FindSubmatch as fields (i.e. it returns the matched expression and groups)
* --matches=pattern: returns 100 if any line matches the pattern, 101 otherwise (while still echoing the input)
* --ofs=c: the character (or string) c is used as output field separator (default space)
* --quote: output fields are quoted
* --printf=format: format output fields according specified format
* output field list: one or more indices (or slices) of fields to return.

also:
* Slices follow the Go slice convention (start:end) or better, the Python slice convention (negative values are offsets from the end, so -1 is the last field).
* Field 0 is the input line (as in "awk") so the various fields are indexed from 1 to the number of fields.
