# API thoughts
* APIs should be hard to misuse
* APIs should be obvious
* APIs should make life easier for the user

## DrawTree
The edgui.DrawTree API is something I'm fairly happy with.  I chose to create
an interface (`TreeNode`) that represents the node of a tree.  This places some
burden on the user - they need to provide four functions (`Name, Children, Leaf, Expanded`).
In return, you get a tree structure draw that supports open/closing of nodes,
any number of children, etc.  Clients likely already have a similar structure 
if there is a tree-like thing already represented.  It feels pretty easy for a
client to create this interface.
In addition, the client can opt into more functionality by implementing richer
interfaces (like `TreeNodeSelected`).
In this model the client is responsible for handling clicks and drags.
The client is also responsible for juggling child nodes.
Another possbile model is to hide the tree node implementation internally and 
carry a client-exposed `any` member.  This doesn't seem like it would reduce
client burden though, and likely widens the API (you need getter/setters for
everything).

## DrawList
A list seems like a much simpler idea.  We have a list (slice) of things we 
wish to show in text form, surely we simply pass a `[]string`?  Let's consider
some scenarios
### Slice of `string`
The simplest API.  Clients must construct a `[]string` and the API will draw it.
Clients will not need to create any new types.  They likely need to transform
one sort of data type to a string.  For just showing lists, `string` is likely
fine.  Trouble begins when the List API needs to handle events.
* Click.  The client will be given an index and the string.  From that data
the client must reverse map back to their original data source.  The reverse
mapping might not be trivial
* Drag to re-order.  The client will need to handle mutation of the original 
list.  Likely not a big deal, the API will inform clients something like "item
n moved to index m"

### Slice of `ListNode`
This is the same method as the Tree.  Clients need to implement an interface
and can then pass a slice of `ListNode` to the `DrawList` function.  Generics
smooth this by not requiring a new slice.  In other words, in many cases 
clients can trivially pass an existing array to `DrawList`.
Event handling is simpler because events are fired on the original data objects.
The API is also extensible by having richer interfaces as needed.