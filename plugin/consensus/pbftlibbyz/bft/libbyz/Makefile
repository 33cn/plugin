all: libbyz.a

CC:=g++
C:=gcc
AR:=ar
RANLIB:=ranlib
MAKEDEPEND:=makedepend -Y
MAKE=gmake
INCS:=-I../gmp -I../sfs/include/sfslite

CPPFLAGS:= $(INCS) -g -Wall -DRECOVERY 
#CPPFLAGS:= $(INCS) -O9 -funroll-loops -ffast-math -malign-double -march=pentiumpro -fomit-frame-pointer -fno-exceptions -DNDEBUG 
#CPPFLAGS:= $(INCS) -O3 -march=pentiumpro -fomit-frame-pointer -fno-exceptions -DNDEBUG   -DRECOVERY
#CPPFLAGS:= $(INCS) -O3 -march=pentiumpro -fno-exceptions -DNDEBUG   -DRECOVERY -pg
#CPPFLAGS:= $(INCS) -O3 -march=pentiumpro -fno-exceptions -DNDEBUG   -DRECOVERY 

CFLAGS:= $(CPPFLAGS)

SINCS:= #-I../sfs  -I../sfs/crypt -I../sfs/async -I../sfs/arpc -I../sfs/sfsmisc -I../sfs/svc -I../sfs/include/sfs 

Principal.o: Principal.cc Principal.h
	$(CC) $(CPPFLAGS) $(SINCS) -o Principal.o -c Principal.cc
Node.o: Node.cc Node.h
	$(CC) $(CPPFLAGS) $(SINCS) -o Node.o -c Node.cc

%.o:: %.cc
	$(CC) $(CPPFLAGS) -o $@ -c $<

%.o:: %.c
	$(C) $(CFLAGS) -o $@ -c $<


C_FILES=\
Client.cc             Replica.cc            New_key.cc\
Commit.cc             Message.cc            Reply.cc \
Digest.cc             Node.cc               Request.cc \
Checkpoint.cc         Pre_prepare.cc        Req_queue.cc \
Prepare.cc            Status.cc             Prepared_cert.cc\
ITimer.cc             Principal.cc          Log_allocator.cc \
Meta_data.cc          Data.cc               Fetch.cc \
Meta_data_cert.cc     State.cc              libbyz.cc \
View_change.cc        New_view.cc           View_change_ack.cc \
View_info.cc          NV_info.cc            Rep_info.cc \
Meta_data_d.cc        Query_stable.cc       Reply_stable.cc \
Stable_estimator.cc   Statistics.cc         Time.cc \
Big_req_table.cc      Pre_prepare_info.cc


c_FILES= fail.c umac.c

H_FILES := $(C_FILES:%.C=%.H)

h_FILES := $(c_FILES:%.c=%.h)

OBJ_FILES:=$(C_FILES:%.cc=%.o) $(c_FILES:%.c=%.o)

objs: $(OBJ_FILES)

clean:
	-rm *.o
	-rm *~

clobber:
	-rm *.o
	-rm libbyz.a
	-rm *~

depend:
	$(MAKEDEPEND) $(INCS) $(SINCS) $(C_FILES)

libbyz.a: objs
	$(AR) rcv $@ $(OBJ_FILES) 
	$(RANLIB) $@

# DO NOT DELETE THIS LINE -- make depend depends on it.
