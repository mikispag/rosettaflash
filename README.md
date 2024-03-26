Rosetta Flash ([CVE-2014-4671](http://web.nvd.nist.gov/view/vuln/detail?vulnId=CVE-2014-4671))
==================================================================================

Adobe Flash Player before 13.0.0.231 and 14.x before 14.0.0.145 on Windows and OS X and before 11.2.202.394 on Linux, Adobe AIR before 14.0.0.137 on Android, Adobe AIR SDK before 14.0.0.137, and Adobe AIR SDK & Compiler before 14.0.0.137 do not properly restrict the SWF file format, which allows remote attackers to conduct cross-site request forgery (CSRF) attacks against JSONP endpoints, and obtain sensitive information, via a crafted OBJECT element with SWF content satisfying the character-set requirements of a callback API.

Slides: https://static.miki.it/RosettaFlash/RosettaFlash.pdf

Writeup: https://blog.miki.it/posts/abusing-jsonp-with-rosetta-flash/

Contact me: https://miki.it/contact

Build instructions
-------------------

To get the code:

``$ go get github.com/mikispag/rosettaflash``

Then, get into ``$GOPATH/github.com/mikispag/rosettaflash`` and use the ``go build`` command to compile.
