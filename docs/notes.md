# Flatland
Flatland is a 2D game/engine for quickly building fun little things.
It should be easy to use and intuitive, but also well designed.

# Table of Contents
* [Decision Log](decisions.md) - why certain decisions were made

# Open problems
- The editor and other systems will be reflection heavy, but typical usage 
for Go reflection is to ignore unexported symbols.  How can I balance out 
information hiding/abstraction against reflection?
 - Note, reflection can actually read all symbols and with some hackery write them, but that's going to require building things like my own JSON encoder
 - Convention might be easiest, ie, for anything you want to save, but be treated as "private", have a naming convention.

# Next Goal
- Tree view widget
- Tree view of components
- Allow inline saving of assets that have parents
 - ie, components have a parent and the actor an override fields
- Create a tiny game that does not include the editor package, but does load the created asset
- Filter asset selection by type
- Clean everything up and decide on how the lines of separation need to be drawn
- Save/restore editor state
- Parse own code to get doc strings on types?

# Done
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

# Parent Values for Assets
Assets may take their default values from another Asset of the *same type*.  This allows for some basic reuse of data values, with the following rules.
1. Assets in memory have the full set of data, there is no connection to their parent at runtime.  In other words, you can change parent values in memory and no already loaded child assets will change.
2. Parent values are applied at Load time, using the values of the in-memory parent asset.  No extra reloading from disk is done.
3. Assets are saved sparsely, only the fields that differ from their parent's field are saved.
4. When an asset is loaded, the zero object for that asset is created, the parent object is copied into the child field by field and then the sparse save data is loaded into the child object.

