# KoboToAnki
Small application to add words saved in the Kobo experimental "My Words" feature to an Anki deck using the Merriam Webster Dictionary API to fetch definitions for the words. 

## Prerequirements
1. Currently the tool only works on MacOS
2. You must have Anki installed, as well as the [AnkiConnect](https://ankiweb.net/shared/info/2055492159) plugin intalled
3. You must have the ["My Words"](https://goodereader.com/blog/kobo-ereader-news/kobo-e-readers-now-save-words-you-looked-up-in-the-dictionary) feature enabled on Kobo
4. Since the tool uses the Merriam Webster API to fetch definitions, you must have a key available. You can get one by getting an account on their [developer website](https://dictionaryapi.com/). Refer to the usage section to see how to inform `koboToAnki` of your API key

## Installation

## Usage
Once you have the tool installed, first set the Merriam API Key. You can do this with the following command
```
koboToAnki -setMerriam <YOU_KEY_HERE>
```
After this, simply running `koboToAnki` while your Kobo reader is connected will add your words to Anki.

If you want to add your words to a specific deck, you can set the deck with the following command
```
koboToAnki -setDeck <DECK_NAME_HERE>
```
## How it works
When the experimental feature "My Words" is enabled on your Kobo, a new table is added to the `.kobo/KoboReader.sqlite` database on your kobo called `WordList` which stores the words you have saved. 
This program extracts the words from that table, then makes request to the Merriam Webster Dictionary API to fetch the definitions.
These definitions are then formatted into Anki cards and sent to Anki via the AnkiConnect API to be added to a specified deck (The default is just the "My Words" deck). In order to do this, `koboToAnki` first launches an instance of Anki and wait for the AnkiConnect server to set up.

# TODO 
- [ ] figure out goreleaser to make easier to install 
