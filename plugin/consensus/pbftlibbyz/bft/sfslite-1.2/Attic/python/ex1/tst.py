
import ex1
import async
import socket

sock = socket.socket ()
fd = sock.fileno ()

x = async.arpc.axprt_stream (fd)

cl = async.arpc.aclnt (x, ex1.foo_prog_1 ())

