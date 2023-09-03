# Open problems
- The editor and other systems will be reflection heavy, but typical usage 
for Go reflection is to ignore unexported symbols.  How can I balance out 
information hiding/abstraction against reflection?

 - Note, reflection can actually read all symbols and with some hackery write them, but that's going to require building things like my own JSON encoder

 - Convention might be easiest, ie, for anything you want to save, but be treated as "private", have a naming convention.

- asset.Load vs asset.NewInstance.  Need to clarify the differences and spell out when to use each.  Improved naming for funcs.

- error handling, or lack of it.  I really want to wrap up every error created and log it at the create point with a stack (?).  Need a way to pipe errors into an editor dialog window

- some sort of global context for the "world", otherwise objects won't be able to interact with each other.

# Next Goal
- Fruitroids game flow - start screen, lives, score, restart
- Custom world editor
- Custom debugger for PIE play
- Asset package needs cleaning, code coverage etc
- Clean everything up and decide on how the lines of separation need to be drawn
- Save/restore editor state
- Parse own code to get doc strings on types?

# Done
- Filter asset selection by type
- Tags on the struct/tag handling clean up (ie `flat:"<tags>"`)
- Create a tiny game that does not include the editor package, but does load the created asset
- Tree view of components on the Actor
- Allow inline saving of assets that have parents
 - ie, components have a parent and the actor an override fields
- Tree view widget
- Editor should update child assets when parent is saved
- PIE
- Figure out reload on asset changed
- Auto load recursive asset paths
- Path selection dialog
- Create a custom asset editor
- Create an asset in the editor


# imgui and context
imgui is lexically scoped and oriented.  It doesn't do super well with
state at a distance from the current scope.  In other words, how do
I pass state around a callstack, like I want to do with the editor funcs
(ie, `structEd`)
The solution I've taken is to pass around a context object (TypeEditorContext)
that can be used to save edit state.  The context is created when an asset edit
window opens and is disposed of when the window closes.

# Components vs Actors
If you only have the Owner node, you also then need all the leaves to generate the tree.  If each node has children, saving the tree is trivial.
