#!/usr/bin/python
import sys

for line in sys.stdin:
    line = line.rstrip()
    if line.isalnum():
        print "OK"
    else:
        print "KO"