
class AsyncException (Exception):
    """Base class for async exceptions."""
    def __init__(self, value):
        self.value = value
    def __str__(self):
        return repr(self.value)

class AsyncXDRException (AsyncException):
    """XDR encode/decode error within async."""
    def __init__(self, value):
        self.value = value
    def __str__(self):
        return repr(self.value)

class AsyncRPCException (AsyncException):
    """RPC program exception within async."""
    def __init__(self, value):
        self.value = value
    def __str__(self):
        return repr(self.value)

class AsyncUnionException (AsyncException,UnboundLocalError):
    """Accessing union member that was not switched to."""
    def __init__(self, value):
        self.value = value
    def __str__(self):
        return repr(self.value)
