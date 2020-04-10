//parser is a package for parsing data that have markers at the start and
//ends of patterns of interest.  Patterns are read from csv files.

package parser

import (
	"fmt"
	"log"
	"strings"
	"unicode/utf8"
)

const order = "order"
const items = "items"
const markers = "markers"
const process = "process"
const attribute = "attribute"
const object = "object"
const eof = "EOF"
const start = 0
const end = 1

//Item is what the using package will get through the channel
type Item struct {
	ItemKey   string
	ItemValue string
}

type lexer struct {
	name       string     //name of the lexer for error reporting
	input      string     //string to be scanned
	start      int        //starting position of the current token
	pos        int        //current position in the string
	width      int        //width of the last rune read
	pattern    [][]string //search pattern from the csv file per specification
	process    []string   //it can be blank "", or the string `process`
	objects    []string   //it can be any of the objects in object field of frame
	attributes []string   //it is the strt of the attribute marker
	attribEnd  []string   //it is the end of the attribute marker
	object     string     //put key here to check for the right end key
	attribute  string     //put key here to check for the right end key
	items      chan Item
}
type stateFn func(*lexer) stateFn

//Element is exported so others can use it as a type for pointer

var lexerData lexer
var l = &lexerData
var searchList []string
var orderPat []string             //order of hierarchy
var itemPat map[string][]string   //objects in each element of the hirearchy
var markerPat map[string][]string //start and end attributes of all search items

//AllNodes is the final output of the lexer

func lexText(l *lexer) stateFn {
	for {
		//searchList is one of the list of the marker patterns
		//item is item in the list that matched the prefix
		ok, item := l.hasPrefix(l.input[l.pos:], searchList)
		if ok {
			if l.pos >= l.start {
				//key is the process, object or attribute of the matched item
				//w is zero (start) if the item is a starting pattern
				//w is one (end) if the item is an ending pattern
				k, key, w := l.findKey(item)
				if !k {
					log.Fatal("No key found in: ", item)
				}
				switch w {
				case start:
					l.processStarts(key)
				case end:
					l.processEnds(key)
				default:
					log.Fatal("Find key returned a bad key: ", key)
				}
				l.advance(item)
				return lexText
			}
		}
		if !l.next() {
			break
		}
	}
	l.emit(eof, "")
	return nil
}

func (l *lexer) processStarts(key string) {
	keyClass, ok := getClass(key)
	if !ok {
		log.Fatal(keyClass)
	}
	switch keyClass {
	case process:
		searchList = l.objects
		return
	case object:
		l.emit("nodeType", key)
		l.object = key
		searchList = l.attributes
		return
	case attribute:
		l.setEnd(key)
		l.attribute = key
		searchList = l.attribEnd
		return
	}
}

func (l *lexer) processEnds(key string) {
	keyClass, ok := getClass(key)
	if !ok {
		log.Fatal(keyClass)
	}
	switch keyClass {
	case process:
		return
	case object:
		l.emit(object, "")
		searchList = l.objects
		l.object = ""
		return
	case attribute:
		searchList = l.attributes
		l.emit(key, l.input[l.start:l.pos])
	}
	return
}

func (l *lexer) run() {
	for state := lexText; state != nil; {
		state = state(l)
	}
	close(l.items)
}

func (l *lexer) next() bool {
	if l.pos >= len(l.input) {
		l.width = 0
		return false
	}
	_, l.width =
		utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += l.width
	return true
}

func (l *lexer) advance(prefix string) {
	l.pos += len(prefix)
	l.start = l.pos
	return
}

func (l *lexer) emit(itemKey string, itemValue string) {
	l.items <- Item{itemKey, itemValue}
}

func (l *lexer) hasPrefix(input string, pat []string) (bool, string) {
	candidate := ""
	for _, item := range pat {
		if strings.HasPrefix(input, item) {
			if len(item) > len(candidate) {
				candidate = item
			}
		}
	}
	if len(candidate) > 0 {
		if ok, _, _ := l.findKey(candidate); !ok {
			return false, ""
		}
		return true, candidate
	}
	return false, ""
}

func getClass(key string) (string, bool) {
	if stringInSlice(itemPat[attribute], key) {
		return attribute, true
	}
	if stringInSlice(itemPat[object], key) {
		return object, true
	}
	if stringInSlice(itemPat[process], key) {
		return process, true
	}
	return fmt.Sprintf("Key not valid in getClass: %s", key), false
}

func stringInSlice(sl []string, st string) bool {
	for _, item := range sl {
		if item == st {
			return true
		}
	}
	return false
}

func (l *lexer) findKey(elem string) (bool, string, int) {
	//No need to do anything special about the start markers, just go
	for key, value := range markerPat {
		if value[start] == elem {
			return true, key, start
		}
		//for stop markers, have to check if we are at the process, object
		//or attribute level, so we look up the key in itemPat.  That is
		//because the end item must be matched agaist the start item.
		if stringInSlice(itemPat[attribute], key) {
			if value[end] == elem && l.attribute == key {
				return true, key, end
			}
		}
		if stringInSlice(itemPat[object], key) {
			if value[end] == elem && l.object == key {
				return true, key, end
			}
		}
		if stringInSlice(itemPat[process], key) {
			if value[end] == elem {
				return true, key, end
			}
		}
	}
	return false, "", 0
}

//this is the firt line in the csv file that defines the order of the hirearchy
func (l *lexer) getOrder() []string {
	for _, pat := range l.pattern {
		if pat[0] == order {
			return pat[1:]
		}
	}
	return []string{}
}

//items are the elements of each item in the hirarchy.  They key is the order
//item and the slice is the list of itmes
func (l *lexer) getItems() map[string][]string {
	var patItem = [][]string{}
	var itemMap = map[string][]string{}
	var ok bool
	//part 1, pick out the items
	for n, pat := range l.pattern {
		if pat[0] == items {
			for i := n + 1; i < len(l.pattern); i++ {
				if pat[0] == markers || i >= len(l.pattern) {
					break
				}
				patItem = append(patItem, l.pattern[i])
			}
			break
		}
	}
	//part 2, pack the items in a map
	patItem = patItem[0 : len(patItem)-2]
	for _, item := range patItem {
		if item[0] == markers {
			break
		}
		if _, ok = itemMap[item[0]]; !ok {
			itemMap[item[0]] = []string{item[1]}
		}
		if ok {
			itemMap[item[0]] = append(itemMap[item[0]], item[1])
		}
	}
	return itemMap
}

//markers are the starting and the ending markers for each item at each
//level of the hirearchy
func (l *lexer) getMarkers() map[string][]string {
	var patItem = [][]string{}
	var itemMap = map[string][]string{}
	var ok bool
	//part 1: pick out the markers
	for n, pat := range l.pattern {
		if pat[0] == markers {
			for i := n + 1; i < len(l.pattern); i++ {
				patItem = append(patItem, l.pattern[i])
			}
			break
		}
	}
	//part 2, pack the items in a map
	for _, item := range patItem {
		if item[0] == items || len(item) == 1 {
			break
		}
		if _, ok = itemMap[item[0]]; !ok {
			itemMap[item[0]] = []string{item[1]}
			itemMap[item[0]] = append(itemMap[item[0]], item[2])
		}
		if ok {
			itemMap[item[0]] = append(itemMap[item[0]], item[1])
			itemMap[item[0]] = append(itemMap[item[0]], item[2])
		}
	}
	return itemMap
}

func (l *lexer) setSearch() {
	for idx, key1 := range orderPat {
		for key2, value2 := range itemPat {
			for key3, value3 := range markerPat {
				cond := key1 == key2 && stringInSlice(value2, key3)
				cond0 := cond && idx == 0
				cond1 := cond && idx == 1
				cond2 := cond && idx == 2

				if cond0 {
					//pick up the top level (process) start and end markers
					l.process = value3
					//just in case object runs off the end
					l.objects = append(l.objects, value3[end])
					//just in case attribute runs off the end
					l.attribEnd = append(l.attribEnd, value3[end])
				}
				if cond1 {
					//pick up objects start and end markers
					l.objects = append(l.objects, value3[start])
					l.objects = append(l.objects, value3[end])
					//just in case attributes don't start or terminate with end attribute
					l.attributes = append(l.attributes, value3[start])
					l.attributes = append(l.attributes, value3[end])
				}
				if cond2 {
					//note that search for start and end attributes are different
					l.attributes = append(l.attributes, value3[start])
				}
			}
		}
	}
}

func (l *lexer) setEnd(en string) {
	l.attribEnd = []string{}
	l.attribEnd = append(l.attribEnd, markerPat[en][end])
	for _, item := range itemPat[object] {
		l.attribEnd = append(l.attribEnd, markerPat[item][end])
		l.attribEnd = append(l.attribEnd, markerPat[item][start])

	}
	return
}

//Lex is the starting point of lexer
func Lex(pattern [][]string, input string) chan Item {
	l := &lexer{
		input:   input,
		pattern: pattern,
		items:   make(chan Item),
	}
	orderPat = l.getOrder()
	itemPat = l.getItems()
	markerPat = l.getMarkers()
	l.setSearch()
	searchList = l.process
	go l.run()
	return l.items
}
