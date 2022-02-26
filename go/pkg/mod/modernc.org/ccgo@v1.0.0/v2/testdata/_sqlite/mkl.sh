set -e
ccgo  --ccgo-watch --ccgo-full-paths --ccgo-struct-checks --ccgo-go -o lemon /home/jnml/tmp/sqlite/tool/lemon.c
cp /home/jnml/tmp/sqlite/tool/lempar.c .
cp /home/jnml/tmp/sqlite/src/parse.y .
rm -f parse.h
./lemon   parse.y
