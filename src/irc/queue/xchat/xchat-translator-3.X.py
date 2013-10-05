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

SERVER_IP = "127.0.0.1"
SERVER_PORT = 4242
BUFFER_SIZE = 1024
TIMER = 100

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

WATCHLIST = {}
ACTIVE_JOBS = 0
TIMEOUT_HOOK = None
CONN = None

class Translator:
	def translate(cls, channel, user, text, tgtLang = DEFAULT_LANG, outgoing = False, srcLang = "auto", tgtTxt = None, echoTxt = None, echo = False, kill = False, read = False):
		global CONN

		request = dict(Outgoing = outgoing, Channel = channel, User = user, Srctxt = text, Srclang = srcLang, Tgttxt = tgtTxt, Tgtlang = tgtLang, EchoTxt = echoTxt, Echo = echo, Kill = kill, Read = read)

		cls.connectToServer()
		
		jsonStr = json.dumps(request).encode("utf-8")
		CONN.send(jsonStr)

		return None
	translate = classmethod(translate)

	def readResults(cls, userdata = None):
		global TIMEOUT_HOOK
		global ACTIVE_JOBS
		global BUFFER

		request = dict(Outgoing = None, Channel = None, User = None, Srctxt = None, Srclang = None, Tgttxt = None, Tgtlang = None, EchoTxt = None, Echo = False, Kill = False, Read = True)
		jsonStr = json.dumps(request).encode("utf-8")

		CONN.send(jsonStr)
		result = json.loads(CONN.recv(BUFFER_SIZE).decode("utf-8"))

		if type(result) == dict:
			if result["Outgoing"]:
				xchat.command("say " + result["Tgttxt"])
			elif result["Srctxt"] != result["Tgttxt"]:
				context = xchat.find_context(channel=result["Channel"])
				str = result["Tgttxt"]
				context.emit_print("Channel Message", "_[%s]" %(result["User"]), str)
				
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

	def checkConnection(cls, status):
		r, w, e = select.select([CONN,], [CONN,], [CONN,], 0)

		if status == "r":
			return r

		if status == "w":
			return w

		if status == "e":
			return e

		return False
	checkConnection = classmethod(checkConnection)

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

def findLangCode(cls, langauge):
	if language is None:
		return None

	if language.upper() in cls.LANGUAGES:
		return LANGUAGES[language.upper()]

	return None

def addTranslationJob(text, targetLang, channel, user, outgoing = False):
	global TIMEOUT_HOOK
	global TIMER
	global ACTIVE_JOBS

	ACTIVE_JOBS += 1
	Translator.translate(channel, user, text, targetLang, outgoing)

	if TIMEOUT_HOOK is None:
		TIMEOUT_HOOK = xchat.hook_timer(TIMER, Translator.readResults)
	return None

def translateIncoming(word, word_eol, userdata):
	global DEFAULT_LANG

	channel = xchat.get_info("channel")
	user = word[0].lower()
	key = channel + " " + user
	chanKey = channel + " " + channel

	if (key in WATCHLIST or chanKey in WATCHLIST) and not user.startswith("_["):
		addTranslationJob(word_eol[1], DEFAULT_LANG, channel, user)

	return xchat.EAT_NONE
xchat.hook_print("Channel Message", translateIncoming)
xchat.hook_print("Channel Msg Hilight", translateIncoming)

def addUser(word, word_eol, userdata):
	global DEFAULT_LANG
	global WATCHLIST

	user = word[1]
	dest = DEFAULT_LANG

	WATCHLIST[xchat.get_info("channel") + " " + user.lower()] = "Test"
	xchat.prnt("Now watching user: " + user)
	return xchat.EAT_ALL
xchat.hook_command("ADDTR", addUser)

def addChannel(word, word_eol, userdata):
	global DFAULT_LANG
	global WATCHLIST

	channel = xchat.get_info("channel")

	WATCHLIST[channel + " " + channel] = "Test"
	xchat.prnt("Now watching channel: " + channel)
	return xchat.EAT_ALL
xchat.hook_command("ADDCHAN", addChannel, help = "/ADDCHAN - adds the current channel to the watch list")

def removeUser(word, word_eol, userdata):
	user = word[1]

	if WATCHLIST.pop(xchat.get_info("channel") + " " + user.lower(), None) is not None:
		xchat.prnt("User %s has been removed from the watch list." %user)

	return xchat.EAT_ALL
xchat.hook_command("RMTR", removeUser, help = "/RMTR <user_nick> - removes user_nick from the watch list for automatic translations.")

def translateAndSay(word, word_eol, userdata):
	addTranslationJob(word_eol[2], word[1], xchat.get_info("channel"), word[0].lower(), True)
	return xchat.EAT_ALL
xchat.hook_command("TRSEND", translateAndSay, help="/TRSEND <dest_lang> <text> - translates the <text> into the <desk_lang> langugage.")

def translate(word, word_eol, userdata):
	addTranslationJob(word_eol[2], word[1], xchat.get_info("channel"), word[0].lower())
	return xchat.EAT_ALL
xchat.hook_command("TR", translate, help="/TR <dest_lang> <text> - translates the <text> into the <desk_lang> langugage.")

def printWatchList(word, word_eol, userdata):
	users = [key.split(' ')[1] for key in WATCHLIST.keys()]

	xchat.prnt("WatchList: %s" %(" ".join(users)))
	return xchat.EAT_ALL
xchat.hook_command("LSUSERS", printWatchList, help = "/LSUSERS - prints out all users on the watch list for automatic translations to the screen locally.")

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

def disableRead(word, word_eol, userdata):
	global TIMEOUT_HOOK

	if TIMEOUT_HOOK is not None:
		xchat.unhook(TIMEOUT_HOOK)
		TIMEOUT_HOOK = None

	return xchat.EAT_ALL
xchat.hook_command("TRDISABLE", disableRead, "/TRDISABLE - disables translations and prevents translations results from being read.")

xchat.prnt("Translator is loaded.")
