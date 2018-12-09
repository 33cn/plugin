/* $Id: sfskeymisc.C 1754 2006-05-19 20:59:19Z max $ */

/*
 *
 * Copyright (C) 1999 David Mazieres (dm@uun.org)
 * Copyright (C) 1999 Michael Kaminsky (kaminsky@lcs.mit.edu)
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

#include "sfsmisc.h"
#include "agentmisc.h"
#include "sfskeymisc.h"
#include "xdr_suio.h"

bool opt_pwd_fd;
vec<int> pwd_fds;

str
myusername ()
{
  if (const char *p = getenv ("USER"))
    return p;
  else if ((p = getlogin ()))
    return p;
  else if (struct passwd *pw = getpwuid (getuid ()))
    return pw->pw_name;
  else
    return NULL;
}

str
sfskeysave (str keyfile, const sfskey *sk, bool excl)
{
  str s;
  str2wstr (s);
  assert (!sk->pwd || sk->cost);
  if (!sk->key->export_privkey (&s, sk->keyname, sk->pwd, sk->cost))
    panic ("Exporting private key failed\n");

  if (!str2file (keyfile, s, 0600, excl))
    return keyfile << ": " << strerror (errno);
  return NULL;
}

str
defkeyname (str user)
{
  if (!user && !(user = myusername ()))
    fatal << "cannot find login name\n";
  if (str host = sfshostname ())
    return user << "@" << host;
  else
    fatal << "cannot find local host name\n";
}

str
sfskeygen_prompt (sfskey *sk, str prompt,
		  bool askname, bool nopwd, bool nokbdnoise)
{
  random_start ();
  rndaskcd ();

  sk->key = NULL;
  if (!sk->keyname)
    sk->keyname = defkeyname ();
  
  if (!prompt)
    prompt = sk->keyname;
  warnx << "Creating new key: " << prompt << "\n";

  if (askname)
    sk->keyname = getline ("       Key Label: ", sk->keyname);
  else
    warnx << "       Key Label: " << sk->keyname << "\n";
  if (nopwd)
    sk->pwd = NULL;
  else if (!(sk->pwd = getpwdconfirm ("Enter passphrase: ")))
    return "Aborted.";
  if (!nokbdnoise)
    rndkbd ();
  random_init ();
  if (!sk->cost)
    sk->cost = sfs_pwdcost;

  return NULL;
}

str
sfskeygen (sfskey *sk, u_int nbits, str prompt,
	   bool askname, bool nopwd, bool nokbdnoise, 
	   bool encrypt, sfs_keytype kt) 
{
  str r = sfskeygen_prompt (sk, prompt, askname, nopwd, nokbdnoise);
  if (r) return r;
  u_char opts = SFS_SIGN;
  if (encrypt) opts |= SFS_ENCRYPT;
  if (!(sk->key = sfscrypt.gen (kt, nbits, opts)))
    return "Could not allocate new keypair.";
  sk->generated = true;
  return NULL;
}

static void
setbool (bool *bp)
{
  *bp = true;
}

void
rndkbd ()
{
  warnx << "\nsfskey needs secret bits with which to"
    " seed the random number generator.\n"
    "Please type some random or unguessable text until you hear a beep:\n";
  bool finished = false;
  if (!getkbdnoise (64, &rnd_input, wrap (&setbool, &finished)))
    fatal << "no tty\n";
  while (!finished)
    acheck ();
}

static void
getstr (str *outp, str in)
{
  *outp = in ? in : str ("");
}
str
getpwd (str prompt)
{
  if (!pwd_fds.empty ()) {
    const int fd = pwd_fds.pop_front ();
    scrubbed_suio uio;
    make_sync (fd);
    while (uio.input (fd) > 0)
      ;
    close (fd);
    if (!uio.resid ())
      fatal ("could not read passphrase from fd %d\n", fd);
    wmstr m (uio.resid ());
    uio.copyout (m, m.len ());
    return m;
  }
  else if (opt_pwd_fd)
    fatal ("bad passphrase\n");
  str out;
  if (!getkbdpwd (prompt, &rnd_input, wrap (getstr, &out)))
    fatal ("no tty\n");
  while (!out)
    acheck ();
  return out;
}
str
getpwdconfirm (str prompt)
{
  str prompt2 (strbuf ("%*s", int (prompt.len ()), "Again: "));
  bool again = pwd_fds.empty ();
  for (;;) {
    str p = getpwd (prompt);
    if (!p.len ())
      return NULL;
    if (!again || p == getpwd (prompt2))
      return p;
    warnx << "Mismatch; try again, or RETURN to quit.\n";
  }
}

str
getline (str prompt, str def)
{
  str out;
  if (!getkbdline (prompt, &rnd_input, wrap (getstr, &out), def))
    fatal ("no tty\n");
  while (!out)
    acheck ();
  return out;
}
