"""sfs.setup.build_xdr

Implements the Distuitils 'build_xdr' command, to run the SFS rpcc
compiler on input .x files to output buildable .C and .h files."""

__revision__ = "$Id: build_xdr.py 872 2005-05-18 21:20:58Z max $"

import os, string
import os.path
from types import *
from distutils.core import Command
from distutils.errors import *
from distutils.sysconfig import customize_compiler
from distutils import log

import sfs.setup.rpcc

class build_xdr (Command):

    description = "copmile .x files into .C/.h files"

    user_options = [
        ('build-xdr', 'b',
         "directory to build .C/.h sources to"),
        ('force', 'f',
         "forcibly build everything (ignore file timestamps)"),
        ('rpcc', 'r',
         "specify the rpcc path"),
        ]

    boolean_options = ['force' ]

    def initialize_options (self):
        self.build_xdr = None
        self.force = False
        self.rpcc = None

    def finalize_options (self):
        self.set_undefined_options ('build',
                                    ('build_xdr', 'build_xdr'),
                                    ('force', 'force'),
                                    ('rpcc', 'rpcc'))
        
        self.xdrmods = self.distribution.xdrmods
        if self.xdrmods:
            self.check_xdr_modules (self.xdrmods)
        self.rpcc_obj = sfs.setup.rpcc.rpcc (self.rpcc)

    def check_xdr_modules (self, mods):
        if type(libraries) is not ListType:
            raise DistutilsSetupError, \
                  "'xdrmods' must be a list of typles"
        
        for mod in mods:
            if type(mod) is not TupleType and len(mod) != 2:
                raise DistutilsSetupError, \
                      "each element of 'xdrmods' must be a 2-tuple"
            if type(lib[0]) is not StringType:
                raise DistutilsSetupError, \
                      "first element of each tuple in 'xdrmods' " + \
                      "must be a string (the library name)"
            if type(lib[1]) is not DictionaryType:
                raise DistutilsSetupError, \
                      "second element of each tuple in 'xdrmods' " + \
                      "must be a dictionary (build info)"

    # check_xdr_modules ()

    def run (self):

        if not self.rpcc:
            return
        
        # Cut and pasted from build_clib.py
        from distutils.ccompiler import new_compiler
        self.compiler = new_compiler(compiler=self.compiler,
                                     dry_run=self.dry_run,
                                     force=self.force)
        customize_compiler(self.compiler)

        if self.include_dirs is not None:
            self.compiler.set_include_dirs(self.include_dirs)
        if self.define is not None:
            # 'define' option is a list of (name,value) tuples
            for (name,value) in self.define:
                self.compiler.define_macro(name, value)
        if self.undef is not None:
            for macro in self.undef:
                self.compiler.undefine_macro(macro)

        self.build_xdrmods (self.xdrmods)
        
    # run()

    def build_xdrmods (self, xdrmods):

        for (mod_name, build_info) in xdrmods:
            sources = build_info.get ('sources')
            if sources is None or type(sources) not in (ListType, TupleType):
                raise DistutilsSetupError, \
                      ("in 'xdrmods' option (library '%s'), " +
                       "'sources' must be present and must be " +
                       "a list of source filenames") % mod_name
            sources = list(sources)

            log.info("building '%s' xdrmod", mod_name)

            
