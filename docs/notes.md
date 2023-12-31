# Open problems
- How to connect code and data?  The UE solution is blueprints, and/or
inheritance.
 - Code registers a named factory func that creates an object
 `Register[T]("Name", func() {} T)`
 - Data can connect to code blocks that match the interface
 - Ha, just register an asset!

- The editor and other systems will be reflection heavy, but typical usage 
for Go reflection is to ignore unexported symbols.  How can I balance out 
information hiding/abstraction against reflection?

- Note, reflection can actually read all symbols and with some hackery write them, but that's going to require building things like my own JSON encoder

- Convention might be easiest, ie, for anything you want to save, but be treated as "private", have a naming convention.

- asset.Load vs asset.NewInstance.  Need to clarify the differences and spell out when to use each.  Improved naming for funcs.

- error handling, or lack of it.  I really want to wrap up every error created and log it at the create point with a stack (?).  Need a way to pipe errors into an editor dialog window

- some sort of global context for the "world", otherwise objects won't be able to interact with each other.

# Next Goal
- Actor/Component bounds interface
- Hide Component Children in editor
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


# World Editing
- `flat.World` is typically the top level container for flat games.
- every actor in the world is saved inline
- inline saved actors usually have parents, but it is not required
- the world creates new instances of every actor at BeginPlay
- world ed needs to track the new instances and provide edit controls for them

