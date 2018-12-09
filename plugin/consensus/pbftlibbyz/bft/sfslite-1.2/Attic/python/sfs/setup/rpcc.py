"""sfs.setup.rpcc

Implements an RPCC compiler abstraction for building Python
RPC/XDR modules for use with SFS."""

__revision__ = "$Id: rpcc.py 873 2005-05-18 22:54:17Z max $"

import os, string
import os.path
from types import *
from distutils.core import Command
from distutils.errors import *
from distutils.sysconfig import customize_compiler
from distutils.dep_util import newer_pairwise, newer_group
from distutils import log
import distutils.spawn

import sfs.setup.local

class rpccompiler:
    """Wrapper class around the SFS C compiler."""

    HEADER=1
    CFILE=2

    def __init__ (self,
                  rpcc=None,
                  dry_run=0,
                  verbose=0):
        self.dry_run = dry_run
        self.verbose = verbose
        self.rpcc = self.find_rpcc (rpcc)

    def spawn (self, cmd):
        distutils.spawn.spawn (cmd, dry_run=self.dry_run, verbose=self.verbose)
    
    def find_rpcc (self, loc):
        executable = 'rpcc'

        if loc is not None:
            if os.path.isfile (loc):
                raise DistutilsSetupError, \
                      "cannot find provided rpcc compiler: " + loc
            else:
                return loc
                
        paths = string.split (os.environ['PATH'], os.pathsep)
        for p in paths + sfs.setup.local.lib:
            f = os.path.join (p, executable)
            if os.path.isfile (f):
                return f
        raise DistutilsSetupError, \
              "cannot find an rpcc compiler"
    
    # find_rpcc ()

    def compile_all (self, sources, output_dir):
        if self.rpcc is None:
            raise DistutilsSetupError, "no rpcc compiler found"
        for s in sources:
            base, ext = os.path.splitext (s)
            if ext != '.x':
                raise DistutilsSetupError, \
                      "file without .x extension: '%s'" % s
            if output_dir:
                dir, fl = os.path.split (s)
                outfile = os.path.join (output_dir, fl)
            else:
                outfile = s

            self.compile_hdr (s, outfile)
            self.compile_cfile (s, outfile)

    def to_y_file (self, xfile, e):
        base, ext = os.path.splitext (xfile);
        if ext != '.x':
            raise DistutilsSetupError, \
                  "file without .x extension: '%s'" % s
        return string.join ([base, e], '.')

    def to_c_file (self, xfile):
        return self.to_y_file (xfile, 'C')
    def to_h_file (self, xfile):
        return self.to_y_file (xfile, 'h')

    # compile_all ()

    def compile_hdr (self, inf, out):
        self.compile (inf, self.to_h_file (out), self.HEADER)

    def compile_cfile (self, inf, out):
        self.compile (inf, self.to_c_file (out) , self.CFILE)

    def compile (self, inf, out, mode):
        # only build it if we have to (based on m-times)
        if distutils.dep_util.newer (inf, out):
            if mode == self.HEADER:
                arg = "-pyh"
            elif mode == self.CFILE:
                arg = "-pyc"
            else:
                raise DistutilsSetupError, \
                      "bad compile mode given to rpcc.compile"
            cmd = ( self.rpcc, arg, '-o', out, inf)
            print string.join (cmd, ' ')
            self.spawn (cmd)

    # compile ()
