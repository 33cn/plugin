/* $Id: rpc_util.c,v 1.4 2006/05/30 00:00:19 dm Exp $ */
/*
 * Sun RPC is a product of Sun Microsystems, Inc. and is provided for
 * unrestricted use provided that this legend is included on all tape
 * media and as a part of the software program in whole or part.  Users
 * may copy or modify Sun RPC without charge, but are not authorized
 * to license or distribute it to anyone else except as part of a product or
 * program developed by the user or with the express written consent of
 * Sun Microsystems, Inc.
 *
 * SUN RPC IS PROVIDED AS IS WITH NO WARRANTIES OF ANY KIND INCLUDING THE
 * WARRANTIES OF DESIGN, MERCHANTIBILITY AND FITNESS FOR A PARTICULAR
 * PURPOSE, OR ARISING FROM A COURSE OF DEALING, USAGE OR TRADE PRACTICE.
 *
 * Sun RPC is provided with no support and without any obligation on the
 * part of Sun Microsystems, Inc. to assist in its use, correction,
 * modification or enhancement.
 *
 * SUN MICROSYSTEMS, INC. SHALL HAVE NO LIABILITY WITH RESPECT TO THE
 * INFRINGEMENT OF COPYRIGHTS, TRADE SECRETS OR ANY PATENTS BY SUN RPC
 * OR ANY PART THEREOF.
 *
 * In no event will Sun Microsystems, Inc. be liable for any lost revenue
 * or profits or other special, indirect and consequential damages, even if
 * Sun has been advised of the possibility of such damages.
 *
 * Sun Microsystems, Inc.
 * 2550 Garcia Avenue
 * Mountain View, California  94043
 */

/*
 * rpc_util.c, Utility routines for the RPC protocol compiler 
 */
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <ctype.h>
#include <unistd.h>
#include "rpc_scan.h"
#include "rpc_parse.h"
#include "rpc_util.h"

#define ARGEXT "argument"

static void printwhere (void);

char curline[MAXLINESIZE];	/* current read line */
char *where = curline;		/* current point in line */
int linenum = 0;		/* current line number */

char *infilename;		/* input filename */

#define NFILES 7
char *outfiles[NFILES];		/* output file names */
int nfiles;

FILE *fout;			/* file pointer of current output */
FILE *fin;			/* file pointer of current input */

list *defined;			/* list of defined things */

/*
 * Reinitialize the world 
 */
void
reinitialize (void)
{
  memset (curline, 0, MAXLINESIZE);
  where = curline;
  linenum = 0;
  defined = NULL;
}

/*
 * string equality 
 */
int
streq (char *a, char *b)
{
  return (strcmp (a, b) == 0);
}

/*
 * find a value in a list 
 */
definition *
findval (list * lst, char *val, int (*cmp) ( /* ??? */ ))
{

  for (; lst != NULL; lst = lst->next) {
    if ((*cmp) (lst->val, val)) {
      return (lst->val);
    }
  }
  return (NULL);
}

/*
 * store a value in a list 
 */
void
storeval (list ** lstp, definition * val)
{
  list **l;
  list *lst;


  for (l = lstp; *l != NULL; l = (list **) & (*l)->next);
  lst = ALLOC (list);
  lst->val = val;
  lst->next = NULL;
  *l = lst;
}

static int
findit (definition * def, char *type)
{
  return (streq (def->def_name, type));
}

static char *
fixit (char *type, char *orig)
{
  definition *def;

  def = (definition *) FINDVAL (defined, type, findit);
  if (def == NULL || def->def_kind != DEF_TYPEDEF) {
    return (orig);
  }
  switch (def->def.ty.rel) {
  case REL_VECTOR:
    return (def->def.ty.old_type);
  case REL_ALIAS:
    return (fixit (def->def.ty.old_type, orig));
  default:
    return (orig);
  }
}

char *
fixtype (char *type)
{
  return (fixit (type, type));
}

char *
stringfix (char *type)
{
  if (streq (type, "string")) {
    return ("wrapstring");
  }
  else {
    return (type);
  }
}

void
ptype (char *prefix, char *type, int follow)
{
  if (prefix != NULL) {
    if (streq (prefix, "enum")) {
      f_print (fout, "enum ");
    }
    else {
      f_print (fout, "struct ");
    }
  }
  if (streq (type, "bool")) {
    f_print (fout, "bool_t ");
  }
  else if (streq (type, "string")) {
    f_print (fout, "char *");
  }
  else {
    f_print (fout, "%s ", follow ? fixtype (type) : type);
  }
}

static int
typedefed (definition * def, char *type)
{
  if (def->def_kind != DEF_TYPEDEF || def->def.ty.old_prefix != NULL) {
    return (0);
  }
  else {
    return (streq (def->def_name, type));
  }
}

int
isvectordef (char *type, relation rel)
{
  definition *def;

  for (;;) {
    switch (rel) {
    case REL_VECTOR:
      return (!streq (type, "string"));
    case REL_ARRAY:
      return (0);
    case REL_POINTER:
      return (0);
    case REL_ALIAS:
      def = (definition *) FINDVAL (defined, type, typedefed);
      if (def == NULL) {
	return (0);
      }
      type = def->def.ty.old_type;
      rel = def->def.ty.rel;
    }
  }
}

char *
locase (char *str)
{
  char c;
  static char buf[100];
  char *p = buf;

  while ((c = *str++)) {
    *p++ = (c >= 'A' && c <= 'Z') ? (c - 'A' + 'a') : c;
  }
  *p = 0;
  return (buf);
}

void
pvname_svc (char *pname, char *vnum)
{
  f_print (fout, "%s_%s_svc", locase (pname), vnum);
}

void
pvname (char *pname, char *vnum)
{
  f_print (fout, "%s_%s", locase (pname), vnum);
}

/*
 * print a useful (?) error message, and then die 
 */
void
error (char *msg)
{
  printwhere ();
  f_print (stderr, "%s, line %d: ", infilename, linenum);
  f_print (stderr, "%s\n", msg);
  crash ();
}

/*
 * Something went wrong, unlink any files that we may have created and then
 * die. 
 */
void
crash (void)
{
  int i;

  for (i = 0; i < nfiles; i++) {
    (void) unlink (outfiles[i]);
  }
  exit (1);
}

void
record_open (char *file)
{
  if (nfiles < NFILES) {
    outfiles[nfiles++] = file;
  }
  else {
    f_print (stderr, "too many files!\n");
    crash ();
  }
}

static char expectbuf[100];
static char *toktostr (tok_kind kind);

/*
 * error, token encountered was not the expected one 
 */
void
expected1 (tok_kind exp1)
{
  s_print (expectbuf, "expected '%s'",
	   toktostr (exp1));
  error (expectbuf);
}

/*
 * error, token encountered was not one of two expected ones 
 */
void
expected2 (tok_kind exp1, tok_kind exp2)
{
  s_print (expectbuf, "expected '%s' or '%s'",
	   toktostr (exp1),
	   toktostr (exp2));
  error (expectbuf);
}

/*
 * error, token encountered was not one of 3 expected ones 
 */
void
expected3 (tok_kind exp1, tok_kind exp2, tok_kind exp3)
{
  s_print (expectbuf, "expected '%s', '%s' or '%s'",
	   toktostr (exp1),
	   toktostr (exp2),
	   toktostr (exp3));
  error (expectbuf);
}

void
tabify (FILE * f, int tab)
{
  while (tab--) {
    (void) fputc ('\t', f);
  }
}


static token tokstrings[] =
{
  {TOK_IDENT, "identifier"},
  {TOK_CONST, "const"},
  {TOK_RPAREN, ")"},
  {TOK_LPAREN, "("},
  {TOK_RBRACE, "}"},
  {TOK_LBRACE, "{"},
  {TOK_LBRACKET, "["},
  {TOK_RBRACKET, "]"},
  {TOK_STAR, "*"},
  {TOK_COMMA, ","},
  {TOK_EQUAL, "="},
  {TOK_COLON, ":"},
  {TOK_SEMICOLON, ";"},
  {TOK_UNION, "union"},
  {TOK_STRUCT, "struct"},
  {TOK_SWITCH, "switch"},
  {TOK_CASE, "case"},
  {TOK_DEFAULT, "default"},
  {TOK_ENUM, "enum"},
  {TOK_TYPEDEF, "typedef"},
  {TOK_INT, "int"},
  {TOK_SHORT, "short"},
  {TOK_LONG, "long"},
  {TOK_HYPER, "hyper"},
  {TOK_UNSIGNED, "unsigned"},
  {TOK_DOUBLE, "double"},
  {TOK_FLOAT, "float"},
  {TOK_CHAR, "char"},
  {TOK_STRING, "string"},
  {TOK_OPAQUE, "opaque"},
  {TOK_BOOL, "bool"},
  {TOK_VOID, "void"},
  {TOK_PROGRAM, "program"},
  {TOK_VERSION, "version"},
  {TOK_EOF, "??????"}
};

static char *
toktostr (tok_kind kind)
{
  token *sp;

  for (sp = tokstrings; sp->kind != TOK_EOF && sp->kind != kind; sp++);
  return (sp->str);
}

static void
printbuf (void)
{
  char c;
  int i;
  int cnt;

#define TABSIZE 4

  for (i = 0; (c = curline[i]); i++) {
    if (c == '\t') {
      cnt = 8 - (i % TABSIZE);
      c = ' ';
    }
    else {
      cnt = 1;
    }
    while (cnt--) {
      (void) fputc (c, stderr);
    }
  }
}

static void
printwhere (void)
{
  int i;
  char c;
  int cnt;

  printbuf ();
  for (i = 0; i < where - curline; i++) {
    c = curline[i];
    if (c == '\t') {
      cnt = 8 - (i % TABSIZE);
    }
    else {
      cnt = 1;
    }
    while (cnt--) {
      (void) fputc ('^', stderr);
    }
  }
  (void) fputc ('\n', stderr);
}

char *
make_argname (char *pname, char *vname)
{
  char *name;

  name = (char *) malloc (strlen (pname) + strlen (vname) + strlen (ARGEXT) + 3);
  if (!name) {
    fprintf (stderr, "failed in malloc");
    exit (1);
  }
  sprintf (name, "%s_%s_%s", locase (pname), vname, ARGEXT);
  return (name);
}

bas_type *typ_list_h;
bas_type *typ_list_t;

void
add_type (int len, char *type)
{
  bas_type *ptr;

  if ((ptr = (bas_type *) malloc (sizeof (bas_type))) == (bas_type *) NULL) {
    fprintf (stderr, "failed in malloc");
    exit (1);
  }

  ptr->name = type;
  ptr->length = len;
  ptr->next = NULL;
  if (typ_list_t == NULL) {
    typ_list_t = ptr;
    typ_list_h = ptr;
  }
  else {
    typ_list_t->next = ptr;
    typ_list_t = ptr;
  }
}

bas_type *
find_type (char *type)
{
  bas_type *ptr;

  ptr = typ_list_h;


  while (ptr != NULL) {
    if (strcmp (ptr->name, type) == 0)
      return (ptr);
    else
      ptr = ptr->next;
  }
  return (NULL);
}

int
nullproc (proc_list * proc)
{
  for (; proc != NULL; proc = proc->next) {
    if (streq (proc->proc_num, "0")) {
      return (1);
    }
  }
  return (0);
}
