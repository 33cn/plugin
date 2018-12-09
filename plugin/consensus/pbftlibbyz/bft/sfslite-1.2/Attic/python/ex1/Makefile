

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

SFSLIB = 	-L$(SFSLIBDIR) -lsfscrypt -L$(SFSLIBDIR) -larpc \
		-L$(SFSLIBDIR) -lasync -L$(SFSLIBDIR) -lpyarpc -lgmp \
		-Xlinker -rpath $(SFSLIBDIR)

SFSLIB_C = 	-L$(SFSLIBDIR) -lsfscrypt -L$(SFSLIBDIR) -larpc \
		-L$(SFSLIBDIR) -lasync  -lgmp -ldmalloc

default:	ex1.so

ex1.C:	ex1.x ex1.h
	$(RPCC) -pyc -o $@ $<

ex1.h:	ex1.x
	$(RPCC) -pyh -o $@ $<

ex1srv: ex1c.o ex1srv.o
	$(CXX) -static ex1c.o ex1srv.o $(SFSLIB_C) -o $@

ex1c.o:	ex1c.C
	$(CXX) -g $(PYINC) $(SFSINC) -o $@ -c ex1c.C

ex1c.C:	ex1.x
	$(RPCC) -h -o ex1c.h $<
	$(RPCC) -c -o ex1c.C $<


%.so: %.o
	$(CXX) -shared $< $(SFSLIB) -o $@

.C.o:
	$(CXX) -g $(PYINC) $(SFSINC) -o $@ -c $< 

clean:	
	rm -f *.o *.pyc *.so *.core *_wrap.cxx ex1.C ex1c.C ex1c.h ex1srv ex1.h
