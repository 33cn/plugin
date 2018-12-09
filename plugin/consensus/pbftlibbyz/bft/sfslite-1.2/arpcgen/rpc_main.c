/* $Id: rpc_main.c,v 1.4 2006/05/30 00:00:18 dm Exp $ */
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
 * rpc_main.c, Top level of the RPC protocol compiler. 
 */

#define RPCGEN_VERSION	"199506"	/* This program's version (year & month) */

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include <unistd.h>
#include <ctype.h>

#include <sys/types.h>
#ifdef __TURBOC__
#define	MAXPATHLEN	80
#include <process.h>
#include <dir.h>
#else
#include <sys/param.h>
#include <sys/file.h>
#endif
#include <sys/stat.h>
#include "rpc_parse.h"
#include "rpc_util.h"
#include "rpc_scan.h"

#define EXTEND	1		/* alias for TRUE */
#define DONT_EXTEND	0	/* alias for FALSE */

static int cppDefined = 0;	/* explicit path for C preprocessor */

struct commandline {
  int cflag;			/* xdr C routines */
  int hflag;			/* header file */
  int tflag;			/* dispatch Table file */
  char *infile;			/* input module name */
  char *outfile;		/* output module name */
};


static char *cmdname;

static char *CPP = PATH_CPP;
static char CPPFLAGS[] = "-C";
static char pathbuf[MAXPATHLEN + 1];
static char *incfile = "arpc.h";

#define ARGLISTLEN	20
#define FIXEDARGS         2

static char *arglist[ARGLISTLEN];
static int argcount = FIXEDARGS;


int nonfatalerrors;		/* errors */
int inetdflag /* = 1 */ ;	/* Support for inetd *//* is now the default */
int pmflag;			/* Support for port monitors */
int logflag;			/* Use syslog instead of fprintf for errors */
int tblflag = 1;		/* Support for dispatch table file */
int callerflag;			/* Generate svc_caller() function */
int compatflag;

#define INLINE 3
/*length at which to start doing an inline */

int doinline = INLINE;		/* length at which to start doing an
				   inline. 3 = default if 0, no
				   xdr_inline code */
int indefinitewait;		/* If started by port monitors, hang till it wants */
int exitnow;			/* If started by port monitors, exit after the call */
int timerflag;			/* TRUE if !indefinite && !exitnow */
int newstyle;			/* newstyle of passing arguments (by value) */
int Cflag = 1;			/* ANSI C syntax */
int tirpcflag = 0;		/* generating code for tirpc, by default */

#ifdef __MSDOS__
static char *dos_cppfile = NULL;
#endif

static void c_output (char *, char *, int, char *);
static void h_output (char *, char *, int, char *);
static void t_output (char *, char *, int, char *);
static void addarg (char *);
static void putarg (int, char *);
static void clear_args (void);
static void checkfiles (char *, char *);
static int parseargs (int, char **, struct commandline *);
static void usage (void);
static void options_usage (void);
void c_initialize (void);


int
main (int argc, char **argv)
{
  struct commandline cmd;

  (void) memset ((char *) &cmd, 0, sizeof (struct commandline));
  clear_args ();
  if (!parseargs (argc, argv, &cmd) || !cmd.infile)
    usage ();

  if (cmd.cflag || cmd.hflag || cmd.tflag)
    checkfiles (cmd.infile, cmd.outfile);
  else
    checkfiles (cmd.infile, NULL);

  if (cmd.cflag) {
    c_output (cmd.infile, "-DRPC_XDR", DONT_EXTEND, cmd.outfile);
  }
  else if (cmd.hflag) {
    h_output (cmd.infile, "-DRPC_HDR", DONT_EXTEND, cmd.outfile);
  }
  else if (cmd.tflag) {
    t_output (cmd.infile, "-DRPC_TBL", DONT_EXTEND, cmd.outfile);
  }
  else {
    /* the rescans are required, since cpp may effect input */
    c_output (cmd.infile, "-DRPC_XDR", EXTEND, "_xdr.c");
    reinitialize ();
    h_output (cmd.infile, "-DRPC_HDR", EXTEND, ".h");
    reinitialize ();
    t_output (cmd.infile, "-DRPC_TBL", EXTEND, "_tbl.c");
  }
#ifdef __MSDOS__
  if (dos_cppfile != NULL) {
    (void) fclose (fin);
    (void) unlink (dos_cppfile);
  }
#endif
  exit (nonfatalerrors);
  /* NOTREACHED */
}

/*
 * add extension to filename 
 */
static char *
extendfile (char *path, char *ext)
{
  char *file;
  char *res;
  char *p;

  if ((file = strrchr (path, '/')) == NULL)
    file = path;
  else
    file++;

  res = alloc (strlen (file) + strlen (ext) + 1);
  if (res == NULL) {
    abort ();
  }
  p = strrchr (file, '.');
  if (p == NULL) {
    p = file + strlen (file);
  }
  (void) strcpy (res, file);
  (void) strcpy (res + (p - file), ext);
  return (res);
}

/*
 * Open output file with given extension 
 */
static void
open_output (char *infile, char *outfile)
{
  if (outfile == NULL) {
    fout = stdout;
    return;
  }

  if (infile != NULL && streq (outfile, infile)) {
    f_print (stderr, "%s: output would overwrite %s\n", cmdname,
	     infile);
    crash ();
  }
  fout = fopen (outfile, "w");
  if (fout == NULL) {
    f_print (stderr, "%s: unable to open ", cmdname);
    perror (outfile);
    crash ();
  }
  record_open (outfile);

}

static void
add_warning (void)
{
  f_print (fout, "/*\n");
  f_print (fout, " * Please do not edit this file.\n");
  f_print (fout, " * It was generated using rpcgen.\n");
  f_print (fout, " */\n\n");
}

/* clear list of arguments */
static void 
clear_args (void)
{
  int i;
  for (i = FIXEDARGS; i < ARGLISTLEN; i++)
    arglist[i] = NULL;
  argcount = FIXEDARGS;
}

/*
 * Open input file with given define for C-preprocessor 
 */
static void
open_input (char *infile, char *define)
{
  int pd[2];

  infilename = (infile == NULL) ? "<stdin>" : infile;
#ifdef __MSDOS__
#define	DOSCPP	"\\prog\\bc31\\bin\\cpp.exe"
  {
    int retval;
    char drive[MAXDRIVE], dir[MAXDIR], name[MAXFILE], ext[MAXEXT];
    char cppfile[MAXPATH];
    char *cpp;

    if ((cpp = searchpath ("cpp.exe")) == NULL
	&& (cpp = getenv ("RPCGENCPP")) == NULL)
      cpp = DOSCPP;

    putarg (0, cpp);
    putarg (1, "-P-");
    putarg (2, CPPFLAGS);
    addarg (define);
    addarg (infile);
    addarg (NULL);

    retval = spawnvp (P_WAIT, arglist[0], arglist);
    if (retval != 0) {
      fprintf (stderr, "%s: C PreProcessor failed\n", cmdname);
      crash ();
    }

    fnsplit (infile, drive, dir, name, ext);
    fnmerge (cppfile, drive, dir, name, ".i");

    fin = fopen (cppfile, "r");
    if (fin == NULL) {
      f_print (stderr, "%s: ", cmdname);
      perror (cppfile);
      crash ();
    }
    dos_cppfile = strdup (cppfile);
    if (dos_cppfile == NULL) {
      fprintf (stderr, "%s: out of memory\n", cmdname);
      crash ();
    }
  }
#else
  RC_INT_IGNORE (pipe (pd));
  switch (fork ()) {
  case 0:
    putarg (0, CPP);
    putarg (1, CPPFLAGS);
    addarg (define);
    addarg (infile);
    addarg ((char *) NULL);
    (void) close (1);
    (void) dup2 (pd[1], 1);
    (void) close (pd[0]);
    execv (arglist[0], arglist);
    perror ("execv");
    exit (1);
  case -1:
    perror ("fork");
    exit (1);
  }
  (void) close (pd[1]);
  fin = fdopen (pd[0], "r");
#endif
  if (fin == NULL) {
    f_print (stderr, "%s: ", cmdname);
    perror (infilename);
    crash ();
  }
}

/*
 * Compile into an XDR routine output file
 */

static void
c_output (char *infile, char *define, int extend, char *outfile)
{
  definition *def;
  char *include;
  char *outfilename;
  long tell;

  c_initialize ();
  open_input (infile, define);
  outfilename = extend ? extendfile (infile, outfile) : outfile;
  open_output (infile, outfilename);
  add_warning ();
  if (infile && (include = extendfile (infile, ".h"))) {
    f_print (fout, "#include \"%s\"\n", include);
    free (include);
    /* .h file already contains rpc/rpc.h */
  }
  else
    f_print (fout, "#include <rpc/rpc.h>\n");
  tell = ftell (fout);
  while ((def = get_definition ())) {
    emit (def);
  }
  write_tables ();
  if (extend && tell == ftell (fout)) {
    (void) unlink (outfilename);
  }
}

void
c_initialize (void)
{

  /* add all the starting basic types */

  add_type (1, "int");
  add_type (1, "long");
  add_type (1, "short");
  add_type (1, "bool");

  add_type (1, "u_int");
  add_type (1, "u_long");
  add_type (1, "u_short");

}

char rpcgen_table_dcl[] = "\
#ifndef _RPCGEN_TABLE_DEFINED_\n\
#define _RPCGEN_TABLE_DEFINED_ 1\n\
struct rpcgen_table_tc {\n\
  char *(*proc) ();\n\
  xdrproc_t xdr_arg;\n\
  unsigned len_arg;\n\
  xdrproc_t xdr_res;\n\
  unsigned len_res;\n\
};\n\
#endif /* !_RPCGEN_TABLE_DEFINED_ */\n";


char *
generate_guard (char *pathname)
{
  char *filename, *guard, *tmp;

  filename = strrchr (pathname, '/');	/* find last component */
  filename = ((filename == 0) ? pathname : filename + 1);
  guard = strdup (filename);
  /* convert to upper case */
  tmp = guard;
  while (*tmp) {
    if (islower (*tmp))
      *tmp = toupper (*tmp);
    tmp++;
  }

  guard = extendfile (guard, "_H_RPCGEN");
  return (guard);
}

/*
 * Compile into an XDR header file
 */

static void
h_output (char *infile, char *define, int extend, char *outfile)
{
  definition *def;
  char *outfilename;
  long tell;
  char *guard;
  list *l;

  open_input (infile, define);
  outfilename = extend ? extendfile (infile, outfile) : outfile;
  open_output (infile, outfilename);
  add_warning ();
  guard = generate_guard (outfilename ? outfilename : infile);

  f_print (fout, "#ifndef _%s\n#define _%s\n\n", guard,
	   guard);

#if 0
  f_print (fout, "#define RPCGEN_VERSION\t%s\n\n", RPCGEN_VERSION);
  f_print (fout, "#include <rpc/rpc.h>\n\n");
#else
  f_print (fout, "#include \"%s\"\n\n", incfile);
#endif

#if 0
  f_print (fout, "#ifdef __cplusplus\n"
"\n"
"#ifndef EXTERN\n"
"#define EXTERN extern \"C\" \n"
"#define EXTERN_DEFINED_BY_%s\n"
"#endif /* !EXTERN */\n"
"#ifndef UNION_NAME\n"
"#define UNION_NAME(name) u\n"
"#define UNION_NAME_DEFINED_BY_%s 1\n"
"#endif /* !UNION_NAME */\n"
"#ifndef CONSTRUCT\n"
"#define CONSTRUCT(Type, type)                                       \\\n"
"struct Type : public type {                                         \\\n"
"  Type () { bzero ((type *) this, sizeof (type)); }                 \\\n"
"  ~Type () { xdr_free ((xdrproc_t) xdr_ ## type,                    \\\n"
"                       (char *) (type *) this); }                   \\\n"
"};\n"
"#define CONSTRUCT_DEFINED_BY_%s\n"
"#endif /* !CONSTRUCT */\n"
"\n"
"#else /* !__cplusplus */\n"
"\n"
"#ifndef EXTERN\n"
"#define EXTERN extern\n"
"#define EXTERN_DEFINED_BY_%s\n"
"#endif /* !EXTERN */\n"
"#ifndef UNION_NAME\n"
"#define UNION_NAME(name) u\n"
"#define UNION_NAME_DEFINED_BY_%s 1\n"
"#endif /* !UNION_NAME */\n"
"#ifndef CONSTRUCT\n"
"#define CONSTRUCT(Type, type)\n"
"#define CONSTRUCT_DEFINED_BY_%s\n"
"#endif /* !CONSTRUCT */\n"
"\n"
"#endif /* !__cplusplus */\n", guard, guard, guard, guard, guard, guard);
#endif

  tell = ftell (fout);
  /* print data definitions */
  while ((def = get_definition ())) {
    print_datadef (def);
  }

  /* print function declarations.  
     Do this after data definitions because they might be used as
     arguments for functions */
  for (l = defined; l != NULL; l = l->next) {
    print_funcdef (l->val);
  }
  if (extend && tell == ftell (fout)) {
    (void) unlink (outfilename);
  }
#if 0
  f_print (fout, "\n#ifdef __cplusplus\n"
	   "}\n"
	   "#endif /* __cplusplus */\n");
#endif

  f_print (fout, "\n");

#if 0
  f_print (fout,
"#ifdef EXTERN_DEFINED_BY_%s\n"
"#undef EXTERN\n"
"#undef EXTERN_DEFINED_BY_%s\n"
"#endif /* EXTERN_DEFINED_BY_%s */\n"
"#ifdef UNION_NAME_DEFINED_BY_%s\n"
"#undef UNION_NAME\n"
"#undef UNION_NAME_DEFINED_BY_%s\n"
"#endif /* UNION_NAME_DEFINED_BY_%s */\n"
"#ifdef CONSTRUCT_DEFINED_BY_%s\n"
"#undef CONSTRUCT\n"
"#undef CONSTRUCT_DEFINED_BY_%s\n"
"#endif /* CONSTRUCT_DEFINED_BY_%s */\n",
	   guard, guard, guard, guard, guard, guard, guard, guard, guard);
#endif

  f_print (fout, "\n#endif /* !_%s */\n", guard);
}

/*
 * generate the dispatch table
 */
static void
t_output (char *infile, char *define, int extend, char *outfile)
{
  definition *def;
  int foundprogram = 0;
  char *outfilename;
  char *incfile;

  open_input (infile, define);
  outfilename = extend ? extendfile (infile, outfile) : outfile;
  incfile = extendfile (infile, ".h");
  open_output (infile, outfilename);

  add_warning ();
  f_print (fout, "#include \"%s\"\n", incfile);

  while ((def = get_definition ())) {
    foundprogram |= (def->def_kind == DEF_PROGRAM);
  }
  if (extend && !foundprogram) {
    (void) unlink (outfilename);
    return;
  }
  write_tables ();
}

/*
 * Add another argument to the arg list
 */
static void
addarg (char *cp)
{
  if (argcount >= ARGLISTLEN) {
    f_print (stderr, "rpcgen: too many defines\n");
    crash ();
    /*NOTREACHED */
  }
  arglist[argcount++] = cp;
}

static void
putarg (int where, char *cp)
{
  if (where >= ARGLISTLEN) {
    f_print (stderr, "rpcgen: arglist coding error\n");
    crash ();
    /*NOTREACHED */
  }
  arglist[where] = cp;

}

/*
 * if input file is stdin and an output file is specified then complain
 * if the file already exists. Otherwise the file may get overwritten
 * If input file does not exist, exit with an error 
 */

static void
checkfiles (char *infile, char *outfile)
{

  struct stat buf;

  if (infile)			/* infile ! = NULL */
    if (stat (infile, &buf) < 0) {
      perror (infile);
      crash ();
    };
#if 0
  if (outfile) {
    if (stat (outfile, &buf) < 0)
      return;			/* file does not exist */
    else {
      f_print (stderr,
	       "file '%s' already exists and may be overwritten\n", outfile);
      crash ();
    }
  }
#endif
}

/*
 * Parse command line arguments 
 */
static int
parseargs (int argc, char **argv, struct commandline *cmd)
{
  int i;
  int j;
  int c;
  char flag[(1 << 8 * sizeof (char))];
  int nflags;

  cmdname = argv[0];
  cmd->infile = cmd->outfile = NULL;
  if (argc < 2) {
    return (0);
  }
  flag['c'] = 0;
  flag['h'] = 0;
  flag['o'] = 0;
  flag['t'] = 0;
  for (i = 1; i < argc; i++) {
    if (argv[i][0] != '-') {
      if (cmd->infile) {
	f_print (stderr, "Cannot specify more than one input file!\n");

	return (0);
      }
      cmd->infile = argv[i];
    }
    else {
      for (j = 1; argv[i][j] != 0; j++) {
	c = argv[i][j];
	switch (c) {
	case 'C':
	  compatflag = 1;
	  break;
	case 'c':
	case 'h':
	case 't':
	  if (flag[c]) {
	    return (0);
	  }
	  flag[c] = 1;
	  break;
	case 'i':
	  if (++i == argc) {
	    return (0);
	  }
	  doinline = atoi (argv[i]);
	  goto nextarg;
	case 'r':
	  if (++i == argc) {
	    return (0);
	  }
	  incfile = argv[i];
	  goto nextarg;
	case 'o':
	  if (argv[i][j - 1] != '-' ||
	      argv[i][j + 1] != 0) {
	    return (0);
	  }
	  flag[c] = 1;
	  if (++i == argc) {
	    return (0);
	  }
	  if (c == 'o') {
	    if (cmd->outfile) {
	      return (0);
	    }
	    cmd->outfile = argv[i];
	  }
	  goto nextarg;
	case 'D':
	  if (argv[i][j - 1] != '-') {
	    return (0);
	  }
	  (void) addarg (argv[i]);
	  goto nextarg;
	case 'Y':
	  if (++i == argc) {
	    return (0);
	  }
	  (void) strcpy (pathbuf, argv[i]);
	  (void) strcat (pathbuf, "/cpp");
	  CPP = pathbuf;
	  cppDefined = 1;
	  goto nextarg;
	default:
	  return (0);
	}
      }
    nextarg:
      ;
    }
  }

  cmd->cflag = flag['c'];
  cmd->hflag = flag['h'];
  cmd->tflag = flag['t'];

  pmflag = 0;			/* set pmflag only in tirpcmode */
  inetdflag = 1;		/* inetdflag is TRUE by default */

  /* check no conflicts with file generation flags */
  nflags = cmd->cflag + cmd->hflag + cmd->tflag;

  if (nflags == 0) {
    if (cmd->outfile != NULL || cmd->infile == NULL) {
      return (0);
    }
  }
  else if (nflags > 1) {
    f_print (stderr, "Cannot have more than one file generation flag!\n");
    return (0);
  }
  return (1);
}

static void
usage (void)
{
  f_print (stderr, "usage:  %s infile\n", cmdname);
  f_print (stderr, "\t%s [-c | -h | -t] [-C] [-o outfile] [infile]\n",
	   cmdname);
  options_usage ();
  exit (1);
}

static void
options_usage (void)
{
  f_print (stderr, "options:\n");
  f_print (stderr, "-c\t\tgenerate XDR routines\n");
  f_print (stderr, "-C\t\tgenerate compatability macros in header\n");
  f_print (stderr, "-Dname[=value]\tdefine a symbol (same as #define)\n");
  f_print (stderr, "-h\t\tgenerate header file\n");
  f_print (stderr, "-i size\t\tsize at which to start generating inline code\n");
  f_print (stderr, "-o outfile\tname of the output file\n");
  f_print (stderr, "-t\t\tgenerate RPC dispatch table\n");
  f_print (stderr, "-Y path\t\tdirectory name to find C preprocessor (cpp)\n");

  exit (1);
}
