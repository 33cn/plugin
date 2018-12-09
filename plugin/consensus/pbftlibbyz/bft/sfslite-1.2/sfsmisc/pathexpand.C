/* $Id: pathexpand.C 3758 2008-11-13 00:36:00Z max $ */

/*
 *
 * Copyright (C) 2001 Michael Kaminsky (kaminsky@lcs.mit.edu)
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
#include "rxx.h"

static str
mk_readlinkres_abs (char *readlinkres, str lookup)
{
  assert (lookup[0] == '/');
  
  if (readlinkres[0] == '/')
    return readlinkres;

  int ix = lookup.len() - 1;
  //skip over trailing slashes
  while (lookup[ix] == '/')
    ix--;
  
  //skip over basename
  while (lookup[ix] != '/')
    ix--;
  
  return strbuf () << str (lookup.cstr (), ix + 1) << readlinkres;
}

static str
slashsfs2sch (str path)
{
  size_t lsfsroot = strlen (sfsroot);
  if (strncmp (sfsroot, path.cstr (), lsfsroot))
    return NULL;

  const char *nosfs = path.cstr () + lsfsroot;
  // skip over slashes after sfsroot
  while (nosfs[0] && nosfs[0] == '/')
    nosfs++;
  
  str s;
  char *firstslash = strchr (nosfs, '/');
  if (firstslash)
    s = strbuf (nosfs, firstslash - nosfs);
  else
    s = strbuf ("%s", nosfs);

  if (sfs_parsepath (s))
    return s;
  else
    return NULL;
}

// static str
// print_indent (int recdepth)
// {
//   str s = strbuf ();
//   while (recdepth--)
//     s = s << " ";
//   return s;
// }

static str
pathexpand (str lookup, str *sch, int recdepth = 0)
{
  char readlinkbuf [PATH_MAX + 1];
  str s;

  struct stat sb;
  while (1) {
//     warn << print_indent (recdepth) << "looking up " << lookup << "\n";

    stat (lookup.cstr (), &sb);
    
    errno = 0;
    if ((s = slashsfs2sch (lookup))) {
//       warn << print_indent (recdepth) << "--------FOUND-----: " 
// 	   << s << "\n";
      *sch = s;
      return lookup;
    }
    else {
      int len = readlink (lookup, readlinkbuf, PATH_MAX);
      if (len < 0) {
// 	warn << print_indent (recdepth) << "readlink of " 
// 	     << lookup << " returned error: " << strerror (errno) << "\n";
	return lookup;
      }
      readlinkbuf[len] = 0;
//       warn << print_indent (recdepth) << "readlink of " << lookup 
// 	   << " returned " << readlinkbuf << "\n";
      lookup = mk_readlinkres_abs (readlinkbuf, lookup);
    }
  }
}

static str
pathiterate (str path, str *sch, 
	     int recdepth = 0, bool firsttime = false)
{
//   warn << print_indent (recdepth) << "iteratepath: " << path << "\n";
  str s = pathexpand (path, sch, recdepth);
  if (path == s && !firsttime)
    return path;

  vec<str> components;
  split (&components, "/", s);
  str newpath ("");
  for (unsigned int i = 1; i < components.size (); i++) {
    newpath = pathiterate (newpath << "/" << components[i], sch,
			   recdepth + 5);
//     warn << print_indent (recdepth) << newpath << "\n";
  }
  return newpath;
}

//for looking up self-certifying hostnames in certprog interface
int
path2sch (str path, str *sch)
{
  *sch = NULL;

  if (!path || !path.len ())
    return ENOENT;
  
  if (sfs_parsepath (path)) {
    *sch = path;
    return 0;
  }

  str lookup;
  if (path[0] == '/')
    lookup = path;
  else
    lookup = strbuf () << sfsroot << "/" << path;

  str result = pathiterate (lookup, sch, 0, true);
  //  warn << "RESULT: " << result << "\n";

  if (errno == EINVAL)
    errno = 0;
//   if (*sch)
//     warn << "RETURNING: " << *sch << "\n";
//   warn << "RETURNING ERROR CODE: " << strerror (errno) << "\n";
  return errno;
}
