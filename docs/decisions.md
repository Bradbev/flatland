# Decision log

### 2D
3D is simply too difficult to create assets and art for.  3D games immediately feel like they need to be bigger and more "AAA".  2D is a restriction that will make projects more manageable.

### Use other libraries/engines
I don't want to build everything from scratch.  I want a simple game up and running ASAP, but the game should be within the structure of the engine.

### Assets first
How to manage asset is one of the most important parts of a game, so let's get that right.  I really like the way that Unreal connects assets to the editor.


# Assets
### JSON
Json is human readable and well supported.  Let's start there and optimize later.  Assets won't contain byte data though, there needs
to be a link to another raw file.

# Editor
### Editor does not depend on the game
Compiling editor with the game isn't needed when the reflection is asset based.