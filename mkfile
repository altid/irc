# Use a go mkfile for Altid in general
</$objtype/mkfile

BIN=/$objtype/bin/alt
TARG=irc

all:V: $O.out

$O.out: $OFILES
	go build -o $target cmd/$TARG/*.go

install:V: $O.out
	mkdir -p $BIN
	cp $O.out $BIN/$TARG
	chmod +x $BIN/$TARG

clean:V:
	rm -f $O.out

uninstall:V:
	rm -f $BIN/TARG

