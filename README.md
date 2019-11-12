# lexer

Update:  The content of the "previous note" (below) are still good.  This
update is document one of the changes I said I will make in the future.
The shared data structure is replaced with a new API.  The Lex function
now returns a channel that sends the data to the user.  The form of this data
is:
type Item struct {
	ItemKey   string
	ItemValue string
}

There are three types of items in the pattern csv file.  They are process,
object, and attribute.  Process delimiters start and stop the lexing.  They
do not cause any symbols to be emitted.  An "EOF" string is emitted as ItemKey
with an empty string when end of input is encountered.  When an object delimiter
is encountered, nodeType is emitted.  For attributes, the first field of the
csv pattern file and the found attribute are emitted.

The inputs to the Lex function have not changed.

Previous note:

This lexer is based on the Rob Pike talk in Australia.  I built a few
specialized lexers for specific needs and then decided to generalize
some of the ideas.  Clearly, more work can be done.

Let's say you have an SVG or some other sort of XML or HTTP or any other
type of delimited text file.  You want to extract some of these delimited
fields and do something with them.  In Rob Pike's example, he wanted to
pick out the template elements and process them.

I have called the delimited fields elements. Elements are part of a larger
entity that I have called an object.  Objects are things like grouped svg items
or process elements in a bpmn file.  The items of interest are in
certain parts of the document.  For example, in a bpmn file, we might only
be interested in the process steps and not in the geometry.  In a svg file,
we might be interested in the geometry but not in the styling.  I have
called this "process", but that is just a word.  No particular meaning is
attached to it other than this is the top-level object.

The delimiters, the objects and the process are specified in an csv file
(though the delimiters in this file are | and not commas).  Commas are way
too popular as parts of element delimiters.  You can see examples of these
here.  
https://github.com/Saied74/lexer-user
They are called pattern.csv and pattern2.csv.  They are self-explanatory
for the most part.

If you have watched the Rob Pike video and looked through his slides,
the code should be fairly easy to read.  The APIs are the following:

The Lex function that takes two inputs.  First is the pattern as [][]string.
There is an example of how to process the csv files to generate this pattern
in the lexer-user code referenced above. Of course, you can use your own
approach.  The second input is the string you want to process
(svg, bpmn, etc.).

The output is obtained from AllNodes which is a []element (element is
not exported by itself).  Element has a field VarParts which is a map
populated by the fields of interest and two fields OutNodes and inNodes
that are used to link the elements together for the cases where linking
is needed (e.g. bpmn data).  Finally, the ToLinks is used to process
the OutNodes and InNodes and actually build the linkage.  These functions
are not a part of the lexer and they are done in the lexer-user.
At some future point, I will pull all these fields out of lexer
and implement them in the user program.  

Finally, the main difference between this and Rob Pike's presentation
is that the when I find start and end of elements, I call a processStarts
and processEnds function that try to process the data in somewhat of a
generic manner.  I will probably add a channel in the future instead
of an exported data structure which would make this much cleaner.
