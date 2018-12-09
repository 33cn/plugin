/* -*-c++-*- */
/* $Id: getkbdnoise.C 1117 2005-11-01 16:20:39Z max $ */

/*
 *
 * Copyright (C) 1998 David Mazieres (dm@uun.org)
 *
 * This program is free software; you can redistribute it and/or
 * modify it under the terms of the GNU General Public License as
 * published by the Free Software Foundation; either version 2, or (at
 * your option) any later version.
 *
 * This program is distributed in the hope that it will be useful, but
 * WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
 * General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software
 * Foundation, Inc., 59 Temple Place, Suite 330, Boston, MA 02111-1307
 * USA
 *
 */


#include "crypt.h"
#include "vec.h"
#include <termios.h>

#ifdef __APPLE__
# define DEV_TTY_BUSTED 1
#endif /* __APPLE__ */

static int
getkbdfd ()
{
#ifdef DEV_TTY_BUSTED
  int fd;
  for (fd = 0; !isatty (fd) ; fd++)
    if (fd >= 2)
      return -1;
  const char *ttypath = ttyname (fd);
  if (!ttypath)
    return -1;
#else /* !DEV_TTY_BUSTED */
  const char *ttypath = "/dev/tty";
#endif /* !DEV_TTY_BUSTED */
  return (open (ttypath, O_RDWR));
}

class kbdinput {
protected:
  const int kbdfd;
private:
  datasink *const dst;
  bool lnext;
  bool tok;
  termios torig;
  termios traw;
  suio outq;

  bool fdreset;

protected:
  bool gotsig;

private:
  bool setraw () { return tcsetattr (kbdfd, TCSAFLUSH, &traw) >= 0; }
  bool setorig () { return tcsetattr (kbdfd, TCSAFLUSH, &torig) >= 0; }
  void readcb ();
  void writecb ();

  bool iserase (cc_t c) { return c == torig.c_cc[VERASE]; }
  bool iskill (cc_t c) { return c == torig.c_cc[VKILL]; }
  bool isreprint (cc_t c) {
#ifdef VREPRINT
    return c == torig.c_cc[VREPRINT];
#else /* !VREPRINT */
    return c == '\022';
#endif /* !VREPRINT */
  }
  bool islnext (cc_t c) {
#ifdef VLNEXT
    return c == torig.c_cc[VLNEXT];
#else /* !LNEXT */
    return c == '\026';
#endif /* !LNEXT */
  }

  virtual void gotch (u_char, bool) {}
  virtual void verase () {}
  virtual void vkill () {}
  virtual void vreprint () {}

protected:
  void reset () {
    if (fdreset)
      return;
    fdreset = true;
    if (tok)
      setorig ();
    if (outq.resid ())
      writecb ();
    if (kbdfd >= 0) {
      fdcb (kbdfd, selread, NULL);
      fdcb (kbdfd, selwrite, NULL);
    }
  }
  void iflush ();
  
public:
  kbdinput (datasink *d)
    : kbdfd (getkbdfd ()), dst (d), lnext (false), tok (false),
      fdreset (false), gotsig (false)
    { err_flush (); }
  virtual ~kbdinput () { reset (); close (kbdfd); }

  bool start ();
  void output (str s);
  void flush () { if (outq.resid ()) writecb (); }
};

void
kbdinput::writecb ()
{
  if (outq.output (kbdfd) < 0)
    fatal ("keyboard (output): %m\n");
  if (!outq.resid ()) {
    fdcb (kbdfd, selwrite, NULL);
    fdcb (kbdfd, selread, wrap (this, &kbdinput::readcb));
  }
}

void
kbdinput::output (str s)
{
  suio_print (&outq, s);
  if (outq.resid ()) {
    fdcb (kbdfd, selread, NULL);
    fdcb (kbdfd, selwrite, wrap (this, &kbdinput::writecb));
  }
}

void
kbdinput::iflush ()
{
  tcflush (kbdfd, TCIFLUSH);

  int n = fcntl (kbdfd, F_GETFL);
  if (n < 0)
    return;
  if (!(n & O_NONBLOCK))
    fcntl (kbdfd, F_SETFL, n|O_NONBLOCK);

  timeval tv;
  tv.tv_sec = 0;
  tv.tv_usec = 100000;
  fdwait (kbdfd, selread, &tv);

  char buf[32];
  while (read (kbdfd, buf, sizeof (buf)) > 0)
    ;
  bzero (buf, sizeof (buf));

  if (!(n & O_NONBLOCK))
    fcntl (kbdfd, F_SETFL, n);
}


bool
kbdinput::start ()
{
  if (kbdfd < 0 || !isatty (kbdfd))
    return false;

  /* Make sure we are in the foreground when getting the original
   * attributes.  Many shells actually leave the terminal in cbreak
   * mode when all jobs are in the background. */
  pid_t pgrp;
  if ((pgrp = tcgetpgrp (kbdfd)) > 0 && pgrp != getpgrp ())
    kill (0, SIGTTOU);
  if (tcgetattr (kbdfd, &torig) < 0) {
    warn ("/dev/tty: %m\n");
    return false;
  }

  traw = torig;
  traw.c_iflag &= ~(IMAXBEL|IGNBRK|BRKINT|PARMRK
		     |ISTRIP|INLCR|IGNCR|ICRNL|IXON);
  // traw.c_oflag &= ~OPOST;
  traw.c_lflag &= ~(ECHO|ECHONL|ICANON|ISIG|IEXTEN);
  traw.c_cflag &= ~(CSIZE|PARENB);
  traw.c_cflag |= CS8;
  traw.c_cc[VMIN] = traw.c_cc[VTIME] = 0;

  if (!setraw ()) {
    setorig ();
    warn ("/dev/tty: %m\n");
    return false;
  }
  tok = true;

  getclocknoise (dst);

  fdcb (kbdfd, selread, wrap (this, &kbdinput::readcb));
  return true;
}

void
kbdinput::readcb ()
{
  struct _isigs { int ch; int sig; };
  static const _isigs isig[] = {
    { VINTR, SIGINT },
    { VQUIT, SIGQUIT },
    { VSUSP, SIGTSTP },
#ifdef VDSUSP
    { VDSUSP, SIGTSTP },
#endif /* VDSUSP */
    { -1, -1 }
  };

  u_char c;
  size_t n = read (kbdfd, &c, 1);
  if (n <= 0) {
    setorig ();
    if (n == 0)
      fatal ("keyboard: EOF (with ICANON clear)\n");
    else
      fatal ("keyboard: %m\n");
  }

  dst->update (&c, 1);
  getclocknoise (dst);

  if (!lnext && c != _POSIX_VDISABLE) {
#ifdef VLNEXT
    if (c == torig.c_cc[VLNEXT]) {
      lnext = true;
      return;
    }
#endif /* VLNEXT */
    for (int i = 0; isig[i].sig > 0; i++)
      if (c == torig.c_cc[isig[i].ch]) {
	setorig ();
	tcflush (kbdfd, TCIFLUSH);
	kill (0, isig[i].sig);
	gotsig = true;
	setraw ();
	getclocknoise (dst);
	vreprint ();
	gotsig = false;
	return;
      }
    if (iserase (c))
      verase ();
    else if (iskill (c))
      vkill ();
    else if (isreprint (c))
      vreprint ();
    else
      goto normal;
    return;
  }

 normal:
  bool olnext = lnext;
  lnext = false;
  gotch (c, olnext);
  c = 0;
}

class kbdnoise : public kbdinput {
  size_t nleft;
  const cbv cb;
  u_char lastchar;

  kbdnoise (size_t keys, datasink *dst, cbv cb)
    : kbdinput (dst), nleft (keys), cb (cb),
      lastchar (0)
    { assert (nleft); }
  void finish () { reset (); (*cb) (); delete this; }

  void vreprint () { output (strbuf ("\r                \r%4u ",
				     (u_int) nleft)); }
  void gotch (u_char c, bool) {
    if (c != lastchar && !--nleft) {
      output ("\007\rDONE\n");
      flush ();
      iflush ();
      finish ();
      return;
    }
    lastchar = c;
    vreprint ();
    iflush ();
  }

public:
  static bool alloc (size_t nkeys, datasink *dst, cbv cb) {
    kbdnoise *kn = New kbdnoise (nkeys + 1, dst, cb);
    if (!kn->start ()) {
      delete kn;
      return false;
    }
    kn->gotch (_POSIX_VDISABLE, false);
    return true;
  }
};

bool
getkbdnoise (size_t nkeys, datasink *dst, cbv noisecb)
{
  return kbdnoise::alloc (nkeys, dst, noisecb);
}

class kbdline : public kbdinput {
  str prompt;
  const bool echo;
  const cbs cb;
  vec<char> pw;

  kbdline (str pr, bool echo, datasink *dst, cbs cb)
    : kbdinput (dst), prompt (pr), echo (echo), cb (cb) {}

  void outputch (u_char c) {
    if (!echo) {
      /* Could conceivably thwart keystroke timing attacks like Dawn
       * Song's by making it harder to distinguish portions of an
       * encrypted login session that correspond to regular typing
       * from those corresponding to passwords (typed with echo turned
       * off). */
      output (" \010");
    }
    else if (c < ' ')
      output (strbuf ("^%c", c + '@'));
    else if (c == '\177')
      output ("^?");
    else
      output (strbuf ("%c", c));
  }
  void verase () {
    if (!pw.size ())
      return;
    char &c = pw.back ();
    if (echo) {
      if (u_char (c) < ' ' || c == '\177')
	output ("\010 \010\010 \010");
      else
	output ("\010 \010");
    }
    c = '\0';
    pw.pop_back ();
  }
  void vkill () { while (pw.size ()) verase (); }
  void vreprint () {
    if (!gotsig)
      output ("\n");
    output (prompt);
    for (size_t i = 0; i < pw.size (); i++)
      outputch (pw[i]);
  }
  void finish () {
    output ("\n");
    flush ();
    wmstr p (pw.size ());
    memcpy (p, pw.base (), pw.size ());
    reset ();
    (*cb) (p);
    delete this;
  }
  void gotch (u_char c, bool lnext) {
    if ((c == '\r' || c == '\n') && !lnext)
      finish ();
    else {
      pw.push_back (c);
      outputch (c);
    }
  }

public:
  ~kbdline () { bzero (pw.base (), pw.size ()); }

  static bool alloc (str prompt, bool echo, datasink *dst, cbs cb,
		     str def = NULL) {
    kbdline *kp = New kbdline (prompt, echo, dst, cb);
    if (!kp->start ()) {
      delete kp;
      return false;
    }
    kp->output (prompt);
    if (def)
      for (size_t i = 0; i < def.len (); i++)
	kp->gotch (def[i], true);
    return true;
  }
};

bool
getkbdpwd (str prompt, datasink *dst, cbs cb)
{
  return kbdline::alloc (prompt, false, dst, cb);
}

bool
getkbdline (str prompt, datasink *dst, cbs cb, str def)
{
  return kbdline::alloc (prompt, true, dst, cb, def);
}
