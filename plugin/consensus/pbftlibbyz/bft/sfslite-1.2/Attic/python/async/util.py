#
# $Id: util.py 866 2005-05-16 17:30:08Z max $
#
import async.core
import signal

__all__ = [ "xdrcp", "fixsignals" ]

def xdrcp(x):
    n = x.__new__ (type (x))
    n.str2xdr (x.xdr2str ())
    return n

def fixsignals ():
    async.core.sigcb (signal.SIGINT, async.core.exit)
    
