all: client bfsmount nfsd utimer newbfs

CC:=g++
AR:=ar
RANLIB:=ranlib
MAKEDEPEND:=makedepend -Y
LIBBYZ:=../libbyz/libbyz.a
INCS:=-I../gmp -I../libbyz -I../sfs/include/sfslite 
CPPFLAGS:= $(INCS) -g -Wall
#CPPFLAGS:= $(INCS) -O3 -mcpu=pentiumpro -march=pentiumpro -fomit-frame-pointer -fno-exceptions -DNDEBUG
#CPPFLAGS:= $(INCS) -O3 -mcpu=pentiumpro -march=pentiumpro -fno-exceptions -DNDEBUG
#CPPFLAGS:= $(INCS) -O3 -mcpu=pentiumpro -march=pentiumpro -fno-exceptions -DNDEBUG -pg
C:=gcc
CFLAGS:= $(CPPFLAGS)
LIBDIRS:= -L../gmp -L../libbyz -L../sfs/lib/sfslite
LIBS:= -lbyz -static -lsfscrypt -lasync -lgmp

%.o:: %.cc
	$(CC) $(CPPFLAGS) -o $@ -c $<

%.o:: %.c
	$(C) $(CFLAGS) -o $@ -c $<

MAKE=gmake

C_FILES=\
client_main.cc bfsmount.cc replica_main.cc \
relay_main.cc server_main.cc utimer.cc newbfs.cc \
fileid_hash.cc fh_hash.cc stats.cc

c_FILES=\
nfsd.c fs.c xdr.c svc.c

H_FILES := $(C_FILES:%.C=%.H)

h_FILES := $(c_FILES:%.c=%.h)

OBJ_FILES:=$(C_FILES:%.cc=%.o) $(c_FILES:%.c=%.o)

objs: $(OBJ_FILES)

clean:
	-rm *.o
	-rm *~
	-rm *.bak

clobber:
	-rm *.o
	-rm *~
	-rm *.bak
	-rm client client_instr bfsmount nfsd norep_nfsd relay
	-rm newbfs nfsdp utimer


depend:
	$(MAKEDEPEND) $(INCS) $(C_FILES) $(c_FILES)


client: ../libbyz/libbyz.a objs
	$(CC) -o client $(CPPFLAGS) $(LIBDIRS) client_main.o ../libbyz/libbyz.a $(LIBS)

bfsmount: objs
	$(CC) -o bfsmount $(CPPFLAGS) bfsmount.o ../libbyz/libbyz.a

nfsd: ../libbyz/libbyz.a objs
	$(CC) -o nfsd $(CPPFLAGS) $(LIBDIRS) nfsd.o replica_main.o fs.o xdr.o svc.o fileid_hash.o fh_hash.o stats.o ../libbyz/libbyz.a $(LIBS)

utimer: utimer.o
	$(CC) -o utimer $(CPPFLAGS) utimer.o

newbfs: newbfs.o
	$(CC) -o newbfs $(CPPFLAGS) newbfs.o

norep_nfsd: objs nfsd.c fs.c
	$(C) -o norep_nfsd $(CPPFLAGS) $(LIBDIRS) -DNO_REPLICATION nfsd.c fs.c server_main.o xdr.o svc.o $(LIBS)

relay: objs 
	$(C) -o relay $(CPPFLAGS) $(LIBDIRS) relay_main.o ../libbyz/libbyz.a


client_instr: ../libbyz/libbyz.a objs
	$(CC) -o client_instr $(CPPFLAGS) -DLOGOPS client_main.cc ../libbyz/libbyz.a 



# DO NOT DELETE THIS LINE -- make depend depends on it.
