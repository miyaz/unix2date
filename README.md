# unix2date
convert unixtime included in STDIN to datetime and output.  
"10 digit seconds” and “13 digits milliseconds” are subject to conversion.  
By default, Convert all unixtime strings.
You can filter within specified time period by using the filter options (-f/-t).
Use -h option for other options description.

1. install (homebrew)

```
brew install miyaz/tap/unix2date
```
if you can't use Homebrew, download the binary from [Releases page](https://github.com/miyaz/unix2date/releases).

or, build and use it as follows.
```
go build -o unix2date main.go
```

2. execute

```
% cat << EOS | unix2date
1496405335 1496403935876
1718530010 1718530070235
EOS
2017-06-02T12:08:55Z 2017-06-02T11:45:35.876Z
2024-06-16T09:26:50Z 2024-06-16T09:27:50.235Z
```

3. execute with the filter option

```
% cat << EOS | unix2date -f 2020-01-01T00:00:00Z
1496405335 1496403935876
1718530010 1718530070235
EOS
2024-06-16T09:26:50Z 2024-06-16T09:27:50.235Z
```

4. show help

```
% unix2date -h
---
Usage:
  unix2date [-s]
  unix2date [-ni] [-f YYYY-mm-ddTHH:MM:SS(.NNN)Z] [-t YYYY-mm-ddTHH:MM:SS(.NNN)Z]
Options:
  -s (--summary)         Output only summary. (this option cannot be used with {-n,-i,-f,-t} options
  -n (--no-convert)      Output unixtime without converting
  -i (--invert-filter)   Invert and output filtered results
  -f (--filter-from) [filter start date (ex. 2024-07-01T00:30:00Z)]
  -t (--filter-to)   [filter end date   (ex. 2024-07-01T01:00:00Z)]
                         Output only lines containing unixtime within specified period
  -qt (--quotations) [characters for quotations (default: `"`)
  -sp (--separators) [characters for separators (default: ` ,\t`)
                         Set characters to detect unixtime
```
