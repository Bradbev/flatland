# UI

- Placement isn't much of an issue
- Containers are probably not difficult

# How to bounds test actors/components
In the main loop, go over every component that has a mouse bounds check callback registered.
flatui.OnEnter(bounder, callback)
OnExit

Each component checks the mouse cursor in their tick fn, getting bounds from their parents