# lexer

This lexer is based based on the Rob Pike talk in Australia.  I have tried
to generalize some of the ideas.  Clearly, more work can be done.

Let's say you have an SVG or somme other sort of XML or HTTP or any other
type of delimited data set.  You want to extract some of these delimited
fields and do something with them.  In Rob Pike's example, he wanted to
pick out the template elements and process them.

I have adopted a terminology.  I have called the delimited fields elements.
Elements are part of a larger entity that I have called an object.  objects
are like grouped svg items or process elements in a bpmn xml file.  The
items of interest are in certain parts of the document.  For example in a
bpmn file, we might only be interested in the process steps and not in the
geometry.  In an svg file, we might be interested in the geometry but not
in the styling.  I have called this "process", but that is just a word here.
No particular meaning attached to it other than this is the top level object.

The delimiters, the objects and the process are recorded in an csv file
(though the delimiters in this file are | and not commas).  Commas are way
too popular as parts of element delimiters.  You can see examples of these
here.  
https://github.com/Saied74/lexer-user
They are called pattern.csv and pattern2.csv.  I think they are self
explanatory for the most part.

If you have watched the Rob Pike video and looked through his slides,
the code should be fairly easy to read through.  The APIs are the following:

The Lex function that takes two inputs.  First is the pattern as [][]string.
There is an example of how to process the csv files to generate this pattern
in the lexer-user code referenced above. Of course, you can use your own
approach as well.  The second input is the string you want to process (svg,
bpmn, etc.).

The Lex function is the only input.

The output is obtained from AllNodes which is a []elements (element is
not exported by itself).  Element has a field VarParts which is a map
populated by the fields of interest and two fields OutNodes and inNodes
that are used to link the elements together for the cases where linking
is needed (e.g. bpmn data).  Finally, the ToLinks is used to process
the OutNodes and InNodes and actually build the linkage.  These functions
are really not a part of the lexer and they are done in the lexer-user.
Maybe at some future point, I will pull all these fields out of lexer
and implement it in the user.  

Finally, the main difference between this and Rob Pike's presentation
is that the when I find start and end of elements, I call a processStarts
and processEnds function that try to process the data in somewhat of a
generic manner.  I will probably add a channel in the future instead
of an exported data structure (shades of Fortran Common).
