all: client server

CC:=g++
AR:=ar
RANLIB:=ranlib
MAKEDEPEND:=makedepend -Y
LIBBYZ:=../libbyz/libbyz.a
INCS:=-I../gmp -I../libbyz -I../sfs/include/sfs 
CPPFLAGS:= $(INCS) -ggdb -Wall  
#CPPFLAGS:= $(INCS) -O3 -march=pentiumpro -fno-exceptions -DNDEBUG 
C:=gcc
CFLAGS:= $(CPPFLAGS)
LIBDIRS:=-L../gmp -L../libbyz -L../sfs/lib/sfslite
LIBS:=-lbyz -lsfscrypt -lasync -static -lgmp

%.o:: %.cc
	$(CC) $(CPPFLAGS) -o $@ -c $<

%.o:: %.c
	$(C) $(CFLAGS) -o $@ -c $<

MAKE=gmake

C_FILES=\
replica_main.cc     simple_client.cc\
client_main.cc      simple_server.cc


c_FILES=\

H_FILES := $(C_FILES:%.C=%.H)

h_FILES := $(c_FILES:%.c=%.h)

OBJ_FILES:=$(C_FILES:%.cc=%.o) $(c_FILES:%.c=%.o)

objs: $(OBJ_FILES)

clean:
	-rm *.o
	-rm *~

clobber:
	-rm *.o
	-rm client server
	-rm *~

depend:
	$(MAKEDEPEND) $(INCS) $(SINCS) $(C_FILES)

client: $(LIBBYZ) client_main.o
	$(CC) -o client $(CPPFLAGS) $(LIBDIRS) client_main.o $(LIBS)

server: $(LIBBYZ) replica_main.o
	$(CC) -o server $(CPPFLAGS) $(LIBDIRS) replica_main.o $(LIBS)


bench: all udp_client udp_server udp_mcast_client udp_mcast_server udp_client_auth udp_server_auth lat_graph_args lat_graph_rets server_rets thr_client thr_manager thr_manager_udp thr_client_udp

# For comparing with byzantine replication:
udp_client: simple_client.o $(LIBBYZ)
	$(CC) -o udp_client $(CPPFLAGS) $(LIBDIRS) simple_client.o -lbyz

udp_server: simple_server.o $(LIBBYZ)
	$(CC) -o udp_server $(CPPFLAGS) $(LIBDIRS) simple_server.o -lbyz

udp_mcast_client: udp_mcast_client.cc
	$(CC) -o udp_mcast_client $(CPPFLAGS) $(LIBDIRS) udp_mcast_client.cc -lbyz

udp_mcast_server: udp_mcast_server.cc
	$(CC) -o udp_mcast_server $(CPPFLAGS) $(LIBDIRS) udp_mcast_server.cc -lbyz

#
# Other targets:
#

# For profiling:
clientp: $(LIBBYZ) client_main.o
	$(CC) -pg -static -o clientp $(CPPFLAGS) $(LIBDIRS) client_main.o $(LIBS) /usr/lib/libc_p.a


serverp: $(LIBBYZ) replica_main.o
	$(CC) -pg -static -o serverp $(CPPFLAGS) $(LIBDIRS) replica_main.o  $(LIBS) /usr/lib/libc_p.a

# Simple benchmark with authentication
udp_client_auth: simple_client.cc $(LIBBYZ)
	$(CC) -o udp_client_auth $(CPPFLAGS)  $(LIBDIRS) -DAUTHENTICATION simple_client.cc -lbyz

udp_server_auth: simple_server.cc $(LIBBYZ)
	$(CC) -o udp_server_auth $(CPPFLAGS)  $(LIBDIRS) -DAUTHENTICATION simple_server.cc -lbyz


udp_client1: simple_client.cc $(LIBBYZ)
	$(CC) -o udp_client1 $(CPPFLAGS)  $(LIBDIRS) simple_client1.cc -lbyz

udp_client2: simple_client2.cc $(LIBBYZ)
	$(CC) -o udp_client2 $(CPPFLAGS)  $(LIBDIRS) simple_client2.cc -lbyz

udp_server2: simple_server2.cc $(LIBBYZ)
	$(CC) -o udp_server2 $(CPPFLAGS)  $(LIBDIRS) simple_server2.cc -lbyz


# Client to compute graph of latency for different arg sizes
lat_graph_args: $(LIBBYZ) lat_graph.cc
	$(CC) -o lat_graph_args $(CPPFLAGS) $(LIBDIRS) lat_graph.cc $(LIBS)

# Client to compute graph of latency for different return sizes
lat_graph_rets: $(LIBBYZ) lat_graph.cc
	$(CC) -o lat_graph_rets -DRETS_GRAPH $(CPPFLAGS) $(LIBDIRS)  lat_graph.cc  $(LIBS)

# Server to compute graph of latency for different return sizes
server_rets: $(LIBBYZ) replica_main.cc
	$(CC) -o server_rets -DRETS_GRAPH $(CPPFLAGS) $(LIBDIRS) replica_main.cc $(LIBS)

# Manger and client to compute throughput graphs
thr_manager: $(LIBBYZ) thr_manager.cc
	$(CC) -o thr_manager $(CPPFLAGS) $(LIBDIRS) thr_manager.cc $(LIBS)

thr_client: $(LIBBYZ) thr_client.cc
	$(CC) -o thr_client $(CPPFLAGS) $(LIBDIRS) thr_client.cc $(LIBS)

# Manger and client to compute throughput graphs for udp service
thr_manager_udp: $(LIBBYZ) thr_manager_udp.cc
	$(CC) -o thr_manager_udp $(CPPFLAGS) $(LIBDIRS) thr_manager_udp.cc $(LIBS)

thr_client_udp: $(LIBBYZ) thr_client_udp.cc
	$(CC) -o thr_client_udp $(CPPFLAGS) $(LIBDIRS) thr_client_udp.cc $(LIBS)

# DO NOT DELETE THIS LINE -- make depend depends on it.
