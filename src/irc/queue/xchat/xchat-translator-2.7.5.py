# -*- coding: utf-8 -*-

__module_name__ = "Xchat Translator"
__module_version__ = "-.-"
__module_description__ = "Performs translations from one language to another"
__module_author__ = "Chuong Ngo"

import xchat
import json
import socket
import select

DEFAULT_LANG = "en"
MAX_ERROR = 3
# Must be either True or False and the capitalization matters.
ECHO = False

SERVER_IP = "127.0.0.1"
SERVER_PORT = 4242
BUFFER_SIZE = 1024
TIMER = 100

ENABLE_UPDATELANG = False

LANGUAGES = {
  'AFRIKAANS' : 'af',
  'ALBANIAN' : 'sq',
  'AMHARIC' : 'am',
  'ARABIC' : 'ar',
  'ARMENIAN' : 'hy',
  'AZERBAIJANI' : 'az',
  'BASQUE' : 'eu',
  'BELARUSIAN' : 'be',
  'BENGALI' : 'bn',
  'BIHARI' : 'bh',
  'BULGARIAN' : 'bg',
  'BURMESE' : 'my',
  'CATALAN' : 'ca',
  'CHEROKEE' : 'chr',
  'CHINESE' : 'zh',
  'CHINESE_SIMPLIFIED' : 'zh-CN',
  'CHINESE_TRADITIONAL' : 'zh-TW',
  'CROATIAN' : 'hr',
  'CZECH' : 'cs',
  'DANISH' : 'da',
  'DHIVEHI' : 'dv',
  'DUTCH': 'nl',
  'ENGLISH' : 'en',
  'ESPERANTO' : 'eo',
  'ESTONIAN' : 'et',
  'FILIPINO' : 'tl',
  'FINNISH' : 'fi',
  'FRENCH' : 'fr',
  'GALICIAN' : 'gl',
  'GEORGIAN' : 'ka',
  'GERMAN' : 'de',
  'GREEK' : 'el',
  'GUARANI' : 'gn',
  'GUJARATI' : 'gu',
  'HEBREW' : 'iw',
  'HINDI' : 'hi',
  'HUNGARIAN' : 'hu',
  'ICELANDIC' : 'is',
  'INDONESIAN' : 'id',
  'INUKTITUT' : 'iu',
  'IRISH' : 'ga',
  'ITALIAN' : 'it',
  'JAPANESE' : 'ja',
  'KANNADA' : 'kn',
  'KAZAKH' : 'kk',
  'KHMER' : 'km',
  'KOREAN' : 'ko',
  'KURDISH': 'ku',
  'KYRGYZ': 'ky',
  'LAOTHIAN': 'lo',
  'LATVIAN' : 'lv',
  'LITHUANIAN' : 'lt',
  'MACEDONIAN' : 'mk',
  'MALAY' : 'ms',
  'MALAYALAM' : 'ml',
  'MALTESE' : 'mt',
  'MARATHI' : 'mr',
  'MONGOLIAN' : 'mn',
  'NEPALI' : 'ne',
  'NORWEGIAN' : 'no',
  'ORIYA' : 'or',
  'PASHTO' : 'ps',
  'PERSIAN' : 'fa',
  'POLISH' : 'pl',
  'PORTUGUESE' : 'pt-PT',
  'PUNJABI' : 'pa',
  'ROMANIAN' : 'ro',
  'RUSSIAN' : 'ru',
  'SANSKRIT' : 'sa',
  'SERBIAN' : 'sr',
  'SINDHI' : 'sd',
  'SINHALESE' : 'si',
  'SLOVAK' : 'sk',
  'SLOVENIAN' : 'sl',
  'SPANISH' : 'es',
  'SWAHILI' : 'sw',
  'SWEDISH' : 'sv',
  'TAJIK' : 'tg',
  'TAMIL' : 'ta',
  'TAGALOG' : 'tl',
  'TELUGU' : 'te',
  'THAI' : 'th',
  'TIBETAN' : 'bo',
  'TURKISH' : 'tr',
  'UKRAINIAN' : 'uk',
  'URDU' : 'ur',
  'UZBEK' : 'uz',
  'UIGHUR' : 'ug',
  'VIETNAMESE' : 'vi',
  'WELSH' : 'cy',
  'YIDDISH' : 'yi'
}
LANG_CODES = dict((v,k) for (k,v) in LANGUAGES.items())

WATCHLIST = {}
CHANWATCHLIST = {}
IGNORELIST = {}
ACTIVE_JOBS = 0
TIMEOUT_HOOK = None
CONN = None

class Translator:
	def translate(cls, channel, user, text, tgtLang = DEFAULT_LANG, echo = False, outgoing = False, srcLang = "auto", tgtTxt = None, echoTxt = None, kill = False, read = False):
		global CONN

		request = dict(Outgoing = outgoing, Channel = channel, User = user, Srctxt = text, Srclang = srcLang, Tgttxt = tgtTxt, Tgtlang = tgtLang, Echotxt = echoTxt, Echo = echo, Kill = kill, Read = read)

		cls.connectToServer()
		
		jsonStr = json.dumps(request).encode("utf-8")
		CONN.send(jsonStr)

		return None
	translate = classmethod(translate)

	def readResults(cls, userdata = None):
		global TIMEOUT_HOOK
		global WATCHLIST
		global IGNORELIST
		global ACTIVE_JOBS

		request = dict(Outgoing = None, Channel = None, User = None, Srctxt = None, Srclang = None, Tgttxt = None, Tgtlang = None, Echotxt = None, Echo = False, Kill = False, Read = True)
		jsonStr = json.dumps(request).encode("utf-8")

		CONN.send(jsonStr)
		result = json.loads(CONN.recv(BUFFER_SIZE).decode("utf-8"))

		key = result["Channel"] + " " + result["User"]
		user = result["User"]

		if type(result) == dict:
			if result["Outgoing"]:
				pruser = "- " + user

				txt = pruser  + result["Tgttxt"]
				xchat.command("say " + txt.encode("utf-8"))

				if ECHO:
					context = xchat.find_context(channel=result["Channel"])
					txt = result["Echotxt"].encode("utf-8")
					context.emit_print("Channel Message", "_[echo]", txt)

				if WATCHLIST is not None and key in WATCHLIST:
					dest, src, cnt = WATCHLIST[key]
					cnt = cnt - 1

					if src == "auto" and ENABLE_DEFAULTLANG:
						src = result["Srclang"]

					WATCHLIST[key] = (dest, src, cnt)
				elif user is not None and user != "" and ENABLE_DEFAULTLANG:
					dest = DEFAULT_LANG
					src = result["Srclang"]
					cnt = 0

					WATCHLIST[key] = (dest, src, cnt)

				pass
			elif result["Srclang"] != result["Tgtlang"] and user is not None and user != "":
				context = xchat.find_context(channel=result["Channel"])
				txt = result["Tgttxt"].encode("utf-8")
				context.emit_print("Channel Message", "_[%s]" %(result["User"]), txt)

				if WATCHLIST is not None and key in WATCHLIST:
					dest, src, cnt = WATCHLIST[key]
					cnt = cnt - 1

					if src == "auto" and ENABLE_DEFAULTLANG:
						src = result["Srclang"]

					WATCHLIST[key] = (dest, src, cnt)
				pass

			if result["Srclang"] == result["Tgtlang"] and user is not None and user != "":
				cnt = 1

				if key in WATCHLIST:
					dest, src, cnt = WATCHLIST[key]
					cnt = cnt + 1
					WATCHLIST[key] = (dest, src, cnt)
				else:
					dest = DEFAULT_LANG
					src = result["Srclang"]
					cnt = 1

					WATCHLIST[key] = (dest, src, cnt)

				if cnt >= MAX_ERROR:
					WATCHLIST.pop(key, None)
					IGNORELIST[key] = (dest, src)
				
			ACTIVE_JOBS -= 1
			
		if ACTIVE_JOBS <= 0:
			xchat.unhook(TIMEOUT_HOOK)
			TIMEOUT_HOOK = None
			
			cls.closeConnection()

		return None
	readResults = classmethod(readResults)

	def connectToServer(cls, ip = SERVER_IP, port = SERVER_PORT):
		global CONN

		if CONN is None:
			CONN = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
			CONN.connect((ip, port))

		return None
	connectToServer = classmethod(connectToServer)

	def closeConnection(cls):
		global CONN
		global ACTIVE_JOBS

		request = dict(Outgoing = None, Channel = None, User = None, Srctxt = None, Srclang = None, Tgttxt = None, Tgtlang = None, EchoTxt = None, Echo = False, Kill = True, Read = False)
		jsonStr = json.dumps(request).encode("utf-8")

		CONN.send(jsonStr)
		CONN = None
		ACTIVE_JOBS = 0
		return None
	closeConnection = classmethod(closeConnection)

def findLangCode(language):
	lang = language.upper()

	if lang in LANGUAGES:
		return LANGUAGES[lang]

	if lang.lower() in LANG_CODES:
		return lang.lower()

	return None

def addTranslationJob(text, targetLang, srcLang, channel, user, echo = False, outgoing = False):
	global TIMEOUT_HOOK
	global TIMER
	global ACTIVE_JOBS

	ACTIVE_JOBS += 1
	Translator.translate(channel, user, text, targetLang, echo, outgoing, srcLang)

	if TIMEOUT_HOOK is None:
		TIMEOUT_HOOK = xchat.hook_timer(TIMER, Translator.readResults)
	return None

def removeUser(key):
	global WATCHLIST
	global IGNORELIST

	if WATCHLIST is not None and WATCHLIST.pop(key, None) is not None:
		xchat.prnt("Removed " + key + " from the watch list.")

	if IGNORELIST is not None and IGNORELIST.pop(key, None) is not None:
		xchat.prnt("Removed " + key + " from the ignore list.")

	return None

def quitRemoveUser(word, word_eol, userdata):
	channel = xchat.get_info("channel")
	user = word[0]

	if user is None:
		return xchat.EAT_NONE

	key = channel + " " + user.lower()
	removeUser(key)

	return xchat.EAT_NONE
xchat.hook_print("Quit", quitRemoveUser)

def kickRemoveUser(word, word_eol, userdata):
	channel = xchat.get_info("channel")
	user = word[1]

	if user is None:
		return xchat.EAT_NONE

	key = channel + " " + user.lower()
	removeUser(key)

	return xchat.EAT_NONE
xchat.hook_print("Kick", kickRemoveUser)

def updateUserNick(word, word_eol, userdata):
	channel = xchat.get_info("channel")
	userOld = word[0]
	userNew = word[1]

	if userOld is None or userNew is None:
		return xchat.EAT_NONE

	userOld = userOld.lower()
	userNew = userNew.lower()
	key = channel + " " + userOld

	if key in WATCHLIST:
		dest, src, cnt = WATCHLIST[key]

		if WATCHLIST.pop(key, None) is not None:
			WATCHLIST[xchat.get_info("channel") + " " + userNew.lower()] = (dest, src, cnt)
			xchat.prnt("Watching " + userNew + ", fomerly " + userOld)
		return xchat.EAT_NONE
xchat.hook_print("Change Nick", updateUserNick)

def translateIncoming(word, word_eol, userdata):
	channel = xchat.get_info("channel")
	user = word[0].lower()
	key = channel + " " + user
	chanKey = channel + " " + channel

	if key in WATCHLIST and not user.startswith("_["):
		dest, src, cnt = WATCHLIST[key]
		addTranslationJob(word_eol[1], dest, src, channel, user)

	if chanKey in CHANWATCHLIST and not user.startswith("_["):
		dest, src = CHANWATCHLIST[chanKey]
		addTranslationJob(word_eol[1], dest, src, channel, user)

	return xchat.EAT_NONE
xchat.hook_print("Channel Message", translateIncoming)
xchat.hook_print("Channel Msg Hilight", translateIncoming)

def translateOutgoing(word, word_eol, userdata):
	if len(word) < 2:
		return xchat.EAT_NONE

	channel = xchat.get_info("channel")
	user = word[0].lower()
	key = channel + " " + user

	if key in WATCHLIST:
		dest, src, cnt = WATCHLIST[key]

		if src != "auto":
			addTranslationJob(word_eol[1], src, dest, channel, user, ECHO, True)

		return xchat.EAT_ALL

	key = key[:-1]

	if key in WATCHLIST:
		dest, src, cnt = WATCHLIST[key]

		if src != "auto":
			addTranslationJob(word_eol[1], src, dest, channel, user, ECHO, True)

		return xchat.EAT_ALL

	return xchat.EAT_NONE
xchat.hook_command('', translateOutgoing, help = "Triggers on all /say commands")

def addUser(word, word_eol, userdata):
	global WATCHLIST

	if len(word) < 2:
		return xchat.EAT_ALL

	user = word[1]
	src = "auto"
	dest = DEFAULT_LANG

	if len(word) > 2 :
		src = findLangCode(word[2])
		
		if src is None:
			xchat.prnt("The specified language is invalid.")
			return xchat.EAT_ALL
		pass

	if len(word) > 3:
		lang = findLangCode(word[3])

		if lang is not None:
			dest = lang
		pass

	key = xchat.get_info("channel") + " " + user.lower()
	WATCHLIST[key] = (dest, src, 0)
	xchat.prnt("Now watching user: " + user + ", source: " + src + ", target: " + dest)
	return xchat.EAT_ALL
xchat.hook_command("ADDTR", addUser, help = "/ADDTR {user} {source_language} {target_language} - adds the specified user to the watchlist.  If {source_language} and/or {target_language} is not specified, then 'auto' will be used for the {source_language} and the DEFAULT_LANG will be used for the {target_language}.")

def addChannel(word, word_eol, userdata):
	global CHANWATCHLIST

	channel = xchat.get_info("channel")

	CHANWATCHLIST[channel + " " + channel] = (DEFAULT_LANG, "auto")
	xchat.prnt("Now watching channel: " + channel)
	return xchat.EAT_ALL
xchat.hook_command("ADDCHAN", addChannel, help = "/ADDCHAN - adds the current channel to the channel watch list")

def addIgnore(word, word_eol, userdata):
	global IGNORELIST

	if len(word) < 2:
		return xchat.EAT_ALL

	channel = xchat.get_info("channel")
	user = word[1]

	IGNORELIST[channel + " " + user] = (DEFAULT_LANG, "auto")
	xchat.prnt("Now ignoring user: " + user)
	return xchat.EAT_ALL
xchat.hook_command("ADDIG", addIgnore, help = "/ADDCHAN {user_nick} - adds the {user_nick} to the ignore list")

def manualRemoveUser(word, word_eol, userdata):
	if len(word) < 2:
		return xchat.EAT_ALL

	user = word[1]

	if user is None:
		return xchat.EAT_ALL

	removeUser(xchat.get_info("channel") + " " + user.lower())
	return xchat.EAT_ALL
xchat.hook_command("RMTR", manualRemoveUser, help = "/RMTR {user_nick} - removes {user_nick} from the watch list for automatic translations.")

def removeChannel(word, word_eol, userdata):
	channel = xchat.get_info("channel")

	if CHANWATCHLIST.pop(channel + " " + channel, None) is not None:
		xchat.prnt("Channel %s has been removed from the watch list." %channel)

	return xchat.EAT_ALL
xchat.hook_command("RMCHAN", removeChannel, help = "/RMCHAN - removes the channel from the channel watch list for automatic translations.")

def removeIgnore(word, word_eol, userdata):
	if len(word) < 2:
		return xchat.EAT_ALL

	user = word[1]

	if IGNORELIST.pop(xchat.get_info("channel") + " " + user.lower(), None) is not None:
		xchat.prnt("User %s has been removed from the ignore list." %user)

	return xchat.EAT_ALL
xchat.hook_command("RMIG", removeIgnore, help = "/RMTR {user_nick} - removes {user_nick} from the ignore list.")

def translateAndSay(word, word_eol, userdata):
	if len(word) < 3:
		return xchat.EAT_ALL

	lang = findLangCode(word[1])

	if lang is None:
		xchat.prnt("Invalid language name or code.  Aborting translation.")
		return xchat.EAT_ALL

	addTranslationJob(word_eol[2], lang, "auto", xchat.get_info("channel"), None, ECHO, True)
	return xchat.EAT_ALL
xchat.hook_command("TRSEND", translateAndSay, help="/TRSEND {dest_lang} {text} - translates the {text} into the {desk_lang} langugage.")

def translate(word, word_eol, userdata):
	addTranslationJob(word_eol[2], word[1], "auto", xchat.get_info("channel"), word[0].lower())
	return xchat.EAT_ALL
xchat.hook_command("TR", translate, help="/TR {dest_lang} {text} - translates the {text} into the {desk_lang} langugage.")

def printWatchList(word, word_eol, userdata):
	xchat.prnt("Printing watch list (nick, channel, src, dest, error count)")

	for key in WATCHLIST.keys():
		channel, user = key.split(' ')
		dest, src, cnt = WATCHLIST[key]

		xchat.prnt("- " + user + " " + channel + " " + src + " " + dest + " " + str(cnt))

	return xchat.EAT_ALL
xchat.hook_command("LSUSERS", printWatchList, help = "/LSUSERS - prints out all users on the watch list for automatic translations to the screen locally.")

def printChanWatchList(word, word_eol, userdata):
	xchat.prnt("Printing channel watch list (nick, channel, src, dest)")

	for key in CHANWATCHLIST.keys():
		channel, user = key.split(' ')
		dest, src = CHANWATCHLIST[key]

		xchat.prnt("- " + user + " " + channel + " " + src + " " + dest)

	return xchat.EAT_ALL
xchat.hook_command("LSCHAN", printChanWatchList, help = "/LSCHAN - prints out all channels on the channel watch list for automatic translations to the screen locally.")

def printIgnoreList(word, word_eol, userdata):
	xchat.prnt("Printing ignore list (nick, channel, src, dest)")

	for key in IGNORELIST.keys():
		channel, user = key.split(' ')
		dest, src = IGNORELIST[key]

		xchat.prnt("- " + user + " " + channel + " " + src + " " + dest)

	return xchat.EAT_ALL
xchat.hook_command("LSIG", printIgnoreList, help = "/LSUSERS - prints out all users on the ignore list.")

def initialize(word, word_eol, userdata):
	global CONN
	global ACTIVE_JOBS
	global WATCHLIST
	global TIMEOUT_HOOK

	if TIMEOUT_HOOK is not None:
		xchat.unhook(TIMEOUT_HOOK)
		TIMEOUT_HOOK = None

	if CONN is not None:
		CONN.close()
		CONN = None

	ACTIVE_JOBS = 0
	WATCHLIST = {}

	xchat.prnt("Translator reinitialized")
	return xchat.EAT_ALL
xchat.hook_command("TRINIT", initialize, help = "/TRINIT - reinitializes the plugin.")

def unload_plugin(userdata):
	global TIMEOUT_HOOK
	global CONN

	if TIMEOUT_HOOK is not None:
		xchat.unhook(TIMEOUT_HOOK)
		TIMEOUT_HOOK = None

	if CONN is not None:
		Translator.closeConnection()

	xchat.prnt("Translator is unloaded.")
	return None
xchat.hook_unload(unload_plugin)

xchat.prnt("Translator is loaded.")
