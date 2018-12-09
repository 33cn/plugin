
#include "ex1c.h"
#include "async.h"
#include "arpc.h"

class ex1srv_t {
public:
  void dispatch (svccb *cb);
  ex1srv_t (int fd) ;
  ptr<axprt_stream> x;
  ptr<asrv> s;
};

ex1srv_t::ex1srv_t (int fd)
{
  tcp_nodelay (fd);
  x = axprt_stream::alloc (fd);
  s = asrv::alloc (x, foo_prog_1, wrap (this, &ex1srv_t::dispatch));
}



void
ex1srv_t::dispatch (svccb *sbp)
{
  if (!sbp) {
    warn << "EOF on socket recevied; shutting down\n";
    delete this;
    return;
  }

  u_int p = sbp->proc ();
  switch (p) {
  case FOO_NULL:
    sbp->reply (NULL);
    break;
  case FOO_FUNC: 
    {
      foo_t *f = sbp->Xtmpl getarg<foo_t> ();
      u_int r = f->xx + f->x.len ();
      warn << "received a foo; x=" << f->x << " & xx=" << f->xx << "\n";
      sbp->replyref (r);
      break;
    }
  case FOO_BAR:
    {
      bar_t *b = sbp->Xtmpl getarg<bar_t> ();
      strbuf buf;
      u_int32_t s = 0;
      for (u_int i = 0; i < b->y.size (); i++) 
	s += b->y[i];
      buf << "length=" << b->y.size () << "; sum=" << s << "\n";
      str x = buf;
      warn << "response: " << x << "\n";
      foo_t foo;
      foo.x = x;
      foo.xx = s;
      sbp->replyref (foo);
      break;
    }
  case FOO_BB:
    {
      bb_t *b = sbp->Xtmpl getarg<bb_t> ();
      sbp->replyref (30);
      break;
    }
  default:
    sbp->reject (PROC_UNAVAIL);
    break;
  }
}

static void
new_connection (int lfd)
{
  sockaddr_in sin;
  socklen_t sinlen = sizeof (sin);
  bzero (&sin, sinlen);
  int newfd = accept (lfd, reinterpret_cast<sockaddr *> (&sin), &sinlen);
  if (newfd >= 0) {
    warn ("accepting connection from %s\n", inet_ntoa (sin.sin_addr));
    vNew ex1srv_t (newfd);
  } else if (errno != EAGAIN) {
    warn ("accept failure: %m\n");
  }
}

static bool
init_server (u_int port)
{
  int fd = inetsocket (SOCK_STREAM, port);
  if (!fd) {
    warn << "cannot allocate TCP port: " << port << "\n";
    return false;
  }
  close_on_exec (fd);
  listen (fd, 200);
  fdcb (fd, selread, wrap (new_connection, fd));
}

int
main (int argc, char *argv[])
{
  init_server (3000);
  amain ();

}
