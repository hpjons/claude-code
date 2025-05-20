from functools import lru_cache
from time import time, sleep

@lru_cache
def printout():
    for i in range(1_000_000):
        print(i, end=' ')

printout()
