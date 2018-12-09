
import async.rpctypes
import async
import ex1
import socket
import sys
import signal

active_srvs = [];

def dispatch (srv, sbp):
    print "in dispatch"
    if sbp.eof ():
        print "EOF !!"
        
        # note: need to do this explicitly, since we have a circular
        # data structure: srv -> dispatch (C) -> dispatch (py) -> srv
        srv.clearcb ();
        active_srvs.remove (srv)
        return
    print "procno=", sbp.proc ()
    #sbp.getarg ().warn ()
    
    if sbp.proc () == ex1.FOO_NULL:
        sbp.reply (None)
        
    elif sbp.proc () == ex1.FOO_BAR:
        bar = sbp.getarg ()
        s = 0
        for i in bar.y:
            print i
            s += i
        f = ex1.foo_t ()
        f.x = 'the sum is ' + `s`
        f.xx = s
        sbp.reply (f)
        
    elif sbp.proc () == ex1.FOO_BB:
        bb = sbp.getarg ()
        r = 0
        if bb.aa == ex1.A2:
            s = 0
            for f in bb.b.foos:
                s += f.xx
            for y in bb.b.bar.y:
                s += y
            r = s
        elif bb.aa == ex1.A1:
            r = bb.f.xx
        else:
            r = bb.i
        res = ex1.foo_opq_t ();
        sbp.reply (r)

    elif sbp.proc () == ex1.FOO_FOOZ:
        arg = sbp.getarg ()
        bytes = arg.b;
        f = ex1.foo_t ()
        f.str2xdr (bytes)
        f.warn ()
        sbp.reply (sbp.getarg ())

    elif sbp.proc () == ex1.FOO_OPQ:
        x = ex1.foo_opq_t ();
        x.c = '4432we00rwersfdqwer';
        sbp.reply (x)
        
    else:
        sbp.reject (async.arpc.PROC_UNAVAIL)

def newcon(sock):
    print "calling newcon"
    nsock, addr = sock.accept ()
    print "accept returned; fd=", `nsock.fileno ()`
    print "new connection accepted", addr
    x = async.arpc.axprt_stream (nsock.fileno (), nsock)
    print "returned from axprt_stream init"
    srv = async.arpc.asrv (x, ex1.foo_prog_1 (),
                           lambda y : dispatch (srv, y))
    print "returned from asrv init"
    active_srvs.append (srv)

port = 3000
if len (sys.argv) > 1:
    port = int (sys.argv[1])

sock = socket.socket (socket.AF_INET, socket.SOCK_STREAM)
sock.bind (('127.0.0.1', port))
sock.listen (10);
async.core.fdcb (sock.fileno (), async.core.selread, lambda : newcon (sock))

async.util.fixsignals ()
async.core.amain ()

