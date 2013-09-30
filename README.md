Translation Server
==================

This is a small program, written in Go, that communicates with translate.google.com in order to carry out translations of text between various languages.  It was built for the specific purpose of working with irc scripts to incorporate translation capabilities into an irc client.  For more details on how to talk to the program with a script or irc plugin, please go to the section titled /<some title/>.

This program comes in two versions.  The no-buffering/no-queuing version does not keep finished translation jobs in a queue, hence the name, opting to just send it back to the client immediately.  This version should be used for clients that can accept the finished jobs as soon as they are done.  The buffering/queue version keeps the finished jobs in a queue, awaiting the client to request them.  This is for clients that are unable to accept the jobs as soon as they are finished (not threaded).

Compiling Instructions
======================
1. Install Go v1.1.2 or higher: http://www.golang.org
2. Download this project and extract it.  The base directory for the extracted program will, from this point on, be referred to as PROJECT_BASE.
3. Set your $GOPATH to PROJECT_BASE/lib.  This is only needed for compiling and the environment variable need not be made persistent.
4. Run "go build" specifying TranslationServer.go as the target.  You can specify where you want the binary file to be made.  TranslationServer.go can be found in either PROJECT_BASE/src/buffering or PROJECT_BASE/src/no-buffering.  Build the version that is right for you.  For example: go build PROJECT_BASE/src/buffering/TranslationServer.go
5.  Run the binary file that was created.

License
=======
The MIT License (MIT)

Copyright (c) 2013 Chuong Ngo

Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the "Software"), to deal in
the Software without restriction, including without limitation the rights to
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
the Software, and to permit persons to whom the Software is furnished to do so,
subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
