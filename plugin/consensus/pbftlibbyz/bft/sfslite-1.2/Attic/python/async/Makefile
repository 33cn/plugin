
ifdef LAB
CXX =		/usr/local/gcc33/bin/g++
PRFX =		/usr/local
SFSTAG	=	dbg33
RPCC =		/disk/rael/max/d33/sfs/rpcc/rpcc
else
CXX =		g++
PRFX =		/usr
SFSTAG =	shdbg
RPCC =		/disk/max/d/sfslite/rpcc/rpcc
endif


PYINC =		-I$(PRFX)/include/python2.3
SFSINCDIR =	/usr/local/include/sfslite/$(SFSTAG)
SFSLIBDIR =	/usr/local/lib/sfslite/$(SFSTAG)

SFSINC =	-I$(SFSINCDIR) -I/usr/local/include

SFSLIB = 	-L$(SFSLIBDIR) -lasync -L$(SFSLIBDIR) -larpc \
		-L$(SFSLIBDIR) -lsfscrypt -L$(SFSLIBDIR) -lpyarpc -lgmp \
		-Xlinker -rpath $(SFSLIBDIR)

default:	all

all:		core.so arpc.so rpctypes.so

.C.o:
	$(CXX) -g $(PYINC) $(SFSINC) -o $@ -c $< 

%.so: %.o
	$(CXX) -shared $< $(SFSLIB) -o $@

clean:	
	rm -f *.o *.pyc *.so *.core *_wrap.cxx
