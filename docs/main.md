# Flatland
Flatland is a 2D game/engine for quickly building fun little things.
It should be easy to use and intuitive, but also well designed.

# Table of Contents
* [Decision Log](decisions.md) - why certain decisions were made

# Next Goal
- Path selection dialog
- Save/restore editor state
- Auto load recursive asset paths? (test it)
- Create a tiny game that does not include the editor package, but does load the created asset
- Clean everything up and decide on how the lines of separation need to be drawn

# Done
- Create a custom asset editor
- Create an asset in the editor


# imgui and context
I have a context and initialization problem.
imgui is lexically scoped and oriented.  It doesn't do super well with
state at a distance from the current scope.  In other words, how do
I pass state around a callstack?
How do I start something the first time through the loop?
How do I know something is done with?