
#
# first attempt at a python-based makefile.  doesn't seem to do any
# good, as far as I can tell.
#
import distutils.ccompiler

sfslib = '/usr/local/lib/sfslite/shdbg'
sfsinc = '/usr/local/include/sfslite/shdbg'

cc = distutils.ccompiler.new_compiler ( verbose = 10 )
cc.add_include_dir ( sfsinc )
cc.add_include_dir ( '/usr/include/python2.3' )
cc.add_library_dir ( sfslib )
cc.add_runtime_library_dir ( sfslib )

for l in ( 'async', 'arpc', 'sfscrypt', 'pyarpc', 'gmp' ) :
    cc.add_library ( l )

cc.compile ( sources = [ 'arpc.C' ],
             output_dir =  '/tmp/sfslite'  ,
             debug = 1 )

cc.link ( distutils.ccompiler.CCompiler.SHARED_OBJECT ,
          [ ' arpc.o ' ] , 'arpc' ,
          build_temp = '/tmp/sfslite' )
    
