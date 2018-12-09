#!/usr/bin/env python

"""\
=========================
Python interface to SFS
=========================

Python interface to the libasync/SFS libraries of David
Mazierers (www.fs.net)

Design goals are:

- use libasync event loop

- python interfaces to RPC libraries and compilers.

pysfs is `Free Software`_.

.. _MySQL: http://www.mysql.com/
.. _`Free Software`: http://www.gnu.org/
.. [PEP-0249] http://www.python.org/peps/pep-0249.html

"""

import os
import sys
from distutils.core import setup
from distutils.extension import Extension
import distutils.sysconfig


name = "SFS-%s" % os.path.basename(sys.executable)
version = "0.1"
extra_objects = []
extra_compile_args = [ '-g' ]

libraries = [ 'async', 'arpc', 'sfscrypt', 'pyarpc', 'gmp' ]
library_dirs = [ '/usr/local/lib/sfslite/shdbg' ]
include_dirs = [ '/usr/local/include/sfslite/shdbg' ]

include_dirs = [ d for d in include_dirs if os.path.isdir (d) ]
library_dirs = [ d for d in library_dirs if os.path.isdir (d) ]

classifiers = """
Development Status :: 5 - Production/Stable
Environment :: Other Environment
License :: OSI Approved :: GNU General Public License (GPL)
Operating System :: MacOS :: MacOS X
Operating System :: OS Independent
Operating System :: POSIX
Operating System :: POSIX :: Linux
Operating System :: Unix
Programming Language :: C++
Programming Language :: Python
Topic :: Libraries
"""

py_inc = [ distutils.sysconfig.get_python_inc(plat_specific=s) for s in [0,1]]

metadata = {
    'name': name,
    'version': version,
    'description': "Python wrappers for SFS/libasync",
    'long_description': __doc__,
    'author': "Max Krohn",
    'author_email': "max@maxk.org",
    'license': "GPL",
    'platforms': "ALL",
    'url': "http://www.maxk.org",
    'download_url': "http://foo.foo.com",
    'classifiers': [ c for c in classifiers.split('\n') if c ],
    'py_modules': [
        "async.util",
        "async.__init__"
        ],
    'ext_modules': [
        Extension( name='async.core',
                   sources=['async/core.C'],
                   include_dirs=include_dirs,
                   library_dirs=library_dirs,
                   libraries=libraries,
                   extra_compile_args=extra_compile_args,
                   extra_objects=extra_objects,
                   runtime_library_dirs=library_dirs,
                   language='c++',
                   ),
        Extension( name='async.rpctypes',
                   sources=['async/rpctypes.C'],
                   include_dirs=include_dirs,
                   library_dirs=library_dirs,
                   libraries=libraries,
                   extra_compile_args=extra_compile_args,
                   extra_objects=extra_objects,
                   runtime_library_dirs=library_dirs,
                   language='c++',
                   ),
        Extension( name='ex1.ex1',
                   sources=['ex1/ex1.C'],
                   include_dirs=include_dirs,
                   library_dirs=library_dirs,
                   libraries=libraries,
                   extra_compile_args=extra_compile_args,
                   extra_objects=extra_objects,
                   runtime_library_dirs=library_dirs,
                   language='c++',
                   )
        ],

    'libraries' :
         [ ( 'pyarpc',
             { 'sources' : [ 'pyarpc/py_' + f + '.C'
                             for f in [ 'gen', 'rpctypes', 'util' ] ],
               'include_dirs': include_dirs + py_inc,
               'library_dirs': library_dirs,
               'libraries': libraries,
               'extra_compile_args': extra_compile_args,
               'extra_objects': extra_objects,
               'runtime_library_dirs': library_dirs,
               'language': 'c++',
               }
             )
           ]
    }

setup(**metadata)
