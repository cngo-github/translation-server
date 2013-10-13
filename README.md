Go Translation Server
=====================

This is a small program, written in Go, that communicates with translate.google.com in order to carry out translations of text between various languages.  It was built for the specific purpose of working with irc scripts to incorporate translation capabilities into an irc client.  For more details on how to talk to the program with a script or irc plugin, please go to the section titled {some title}.

This program comes in two versions.  The no-queue version does not keep finished translation jobs in a queue, hence the name, opting to just send it back to the client immediately.  This version should be used for clients that can accept the finished jobs as soon as they are done.  The queue version keeps the finished jobs in a queue, awaiting the client to request them.  This is for clients that are unable to accept the jobs as soon as they are finished (not threaded).

Notes
=====
- All translation jobs are run on their own goroutine that is created for that job.  There is not cap on this on how many goroutines to make in this program.  Be mindful of the number of requests to send to this program and the load it will have on your machine.

Compiling Instructions
======================
1. Install Go v1.1.2 or higher: http://www.golang.org
2. Download this project and extract it.  The base directory for the extracted program will, from this point on, be referred to as PROJECT_BASE.
3. Run "go build" specifying TranslationServer.go as the target.  You can specify where you want the binary file to be made.  TranslationServer.go can be found in either PROJECT_BASE/src/queue or PROJECT_BASE/src/no-queue.  Build the version that is right for you.  For example: go build PROJECT_BASE/src/queue/TranslationServer.go
4. Run the binary file that was created.

XChat/Hexchat IRC Client Plugin
===============================

The two scripts are provided as is and work with the Translation Server program to provide translation capabilities for XChat/Hexchat.  There are two versions of the script.  The "base" version is written in Python 3.X and the IRC client will need a plugin that supports that to run it.  The other version needs Python 2.7.5 or higher, up to Python 3.0.  Both scripts will provide the same functionality.

NOTE: Due to the limitations of XChat, the script polls the Translation Server for finished jobs instead of doing it in a seperate thread or with an event.  By default, the interval for the polling is a tenth of a second.

Setup
=====
Because Xchat/Hexchat does not support threading/forking of plugins, the scripts will use the "queue" version of the translation server.  Below are the steps to compile and run the that version of the translation server.

1. Install Go v1.1.2 or higher: http://www.golang.org
2. Download this project and extract it.  The base directory for the extracted program will, from this point on, be referred to as PROJECT_BASE.
3. Run "go build" specifying TranslationServer.go as the target.  You can specify where you want the binary file to be made.  TranslationServer.go can be found in PROJECT_BASE/src/queue.  For example: go build PROJECT_BASE/src/queue/TranslationServer.go
4. Run the binary file that was created.

After the server is running, load the appropriate script into your IRC client.

1. Open the "Plugins and Scripts" window under Window in the menu bar: Window -> Plugins and Script
2. Look for the Python plugin.  It is normally loaded with XChat/Hexchat at startup.  In the description or the version field, see what version of Python is supports and make note of that.  Hexchat usually comes with a Python 3 plugin while XChat has a Python 2.7.5 plugin.
3. Select the "Load Plugin or Script" from the XChat/Hexchat menu: "XChat -> Load Plugin or Script" or "Hexchat -> Load Plugin or Script"
4. Select the appropriate Python script to load based on the supported Python version you found in step 2.
5. If the plugin loads succesfully, you'll see "Translator is loaded." Printed out in your IRC Client.

Commands
========
Here are the current list of commands:

- /ADDTR {user} {source_language} {target_language} - adds the specified user to the watch list.  If {source_language} and/or {target_language} is not specified, then 'auto' will be used for the {source_language} and the DEFAULT_LANG will be used for the {target_language}.
- /ADDTR {user} - adds the specified user to the ignore list.
- /ADDCHAN - adds the current channel to the watch list
- /RMTR {user_nick} - Removes {user_nick} from the watch list for automatic translations.
- /RMCHAN - removes the current channel from the channel watch list for automatic translations.
- /RMIG {user_nick} - Removes {user_nick} from the watch list for automatic translations.
- /TRSEND {dest_lang} {text} - translates the {text} into the {dest_lang} language and sends the translation to the current channel.
- /TR {dest_lang} {text} - translates the {text} into the {dest_lang} language and prints it locally.
- /LSUSERS - prints the contents of the watch list for automatic translations to the screen locally.
- /LSCHAN - prints out all channels on the channel watch list for automatic translations to the screen locally.
- /LSIG - prints out all users on the ignore list.
- /TRINIT - reinitializes the plugin.

Modifying the Python Script
===========================
- By default, the script will translates all text into English.  If you wish to change this, open up the Python script you'll be using and go to line 13.  Here, change the value of DEFAULT_LANG to the language code of whatever language you desire that is supported.  The list of supported languages and their language codes are from line 25 to line 116.
- By default, if translations fail for a user over and over again, the script will start ignoring that user.  On line 14, the MAX_ERROR value determines the distrust value that users must hit or exceed before being ignored.
- Due to limitations inherent in XChat and Hexchat, the script is not threaded or forked.  Instead, it polls the server on regular intervals to see if there are any finished translation jobs.  The interval is set on line 21, the value of TIMER.  By default, the script polls the server ever tenth of a second.  A TIMER value of 1000 will have the script poll the server every 1 second.
- If you wish to enable echoing, you can do so on line 16, ECHO.  With echoing enabled, two translations will run for each message.  The first will translate the message into the target language, while the second translates it back to the source language.  This provides an "estimate" of what the translation says.  The value of ECHO must either be True or False and the capitalization matters.

Using the XChat/Hexchat Script
==============================
Once the script is loaded, you can use the /TR and /TRSEND commands to translate messages.  In order enable automatic translations, you need to add the user to the watch list using the /ADDTR command.  Once the user has been added, the script will translate all messages incoming from the user.  The translations will be printed to the screen with the same username wrapped in square brackets and an underscore in front.  For example:

- I've added the user randomUser to the watch list with the defaults values for source and target languages.  A message from that user arrives:
- 	randomUser: Some message
-
- The translation will print out as:
-	_[randomuser]: the translated message

If a source language was specified when the user was added to watch list, then automatic translations for outgoing messages addressed to that user will also be done.  For example:

- I've added the user randomUser with the source language fr to the watch list.  I type a message to the user such as:
-	randomUser, some message
-
- The translation will be run and then sent the channel like:
-	- randomUser, un certain message

Additionally, if a user on the watch list quits or gets kicked from the channel, that user will be removed from the watch list automatically.  If the user changes his or her nick, then the watch list will be updated as well.

If you want to translate everything in a channel, use /ADDCHAN to add the channel to the watch list.  Automatic translations will only happen for incoming messages and not outgoing messages, unless there is also an entry in the watchlist for the user.  Be careful to not overload the script with translation requests.

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
