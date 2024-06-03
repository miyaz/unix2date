# unix2date
convert unixtime included in STDIN to datetime and output.  
"10 digit seconds” and “13 digits milliseconds” are subject to conversion.  
By default, datetime within 365 days before and after the current time will be converted. This can be changed with the -days-ago, -days-later options (number of days).

1. build

```
go build -o unix2date main.go
```

2. execute

```
% cat << EOS | ./unix2date
1496405335 1496403935876
1718530010 1718530070235
EOS
1496405335 1496403935876
2024-06-16T09:26:50Z 2024-06-16T09:27:50.235Z
```

3. execute with the period past 10 years

```
% cat << EOS | ./unix2date -days-ago 3650
1496405335 1496403935876
1718530010 1718530070235
EOS
2017-06-02T12:08:55Z 2017-06-02T11:45:35.876Z
2024-06-16T09:26:50Z 2024-06-16T09:27:50.235Z
```

