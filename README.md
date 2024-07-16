# unix2date
convert unixtime included in STDIN to datetime and output.  
"10 digit seconds” and “13 digits milliseconds” are subject to conversion.  
By default, Convert all unixtime strings. You can filter within specified time period by using the filter options (-f/-t).

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
2017-06-02T12:08:55Z 2017-06-02T11:45:35.876Z
2024-06-16T09:26:50Z 2024-06-16T09:27:50.235Z
```

3. execute with the filter option

```
% cat << EOS | ./unix2date -f 2020-01-01T00:00:00Z
1496405335 1496403935876
1718530010 1718530070235
EOS
2024-06-16T09:26:50Z 2024-06-16T09:27:50.235Z
```

