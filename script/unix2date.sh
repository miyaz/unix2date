# シェルで作るとしたら

_unix2date () {
  line="$(cat -)"
  echo "$line" | egrep -q "(^|[^0-9])1[0-9]{9,12}($|[^0-9])"
  if [ $? -ne 0 ]; then
    echo "$line"
  else
    while read LINE
    do
      echo "$LINE" | egrep -q "(^|[^0-9])1[0-9]{9,12}($|[^0-9])"
      if [ $? -eq 0 ]; then
        line=$(echo "$line" | sed "s/${${LINE//[^0-9]/}:0:13}/`date -u "+%FT%TZ" -d @${${LINE//[^0-9]/}:0:10}`/g")
      fi
    done < <(echo "$line" | sed -E "s/[^0-9]+/\n/g")
    echo "$line"
  fi
}

unix2date () {
  data=''
  if [ -t 0 ] ; then
    data="$1"
  else
    data="$(cat -)"
  fi
  LINE_NUM=$(echo "$data" | wc -l)
  if [ ${LINE_NUM} -eq 1 ]; then
    echo "$data" | _unix2date
  else
    echo "$data" |\
      while IFS= read -r line
      do
        echo "$line" | _unix2date
      done
  fi
}

