# Binmap
I've found that using the stdlib `binary` interface to read and write data is a little cumbersome and tedious, since any operation can result in an error.
While this makes sense given the problem domain, the API leaves something to be desired.

I'd love to have a way to batch operations, so I don't have so much `if err != nil`.
If an error occurs at any point, then I'm able to handle one error at the end.

I'd also like to work easily with `io.Reader`s rather than having to read everything into memory first.
While this *can* be accomplished with `binary.Read`, I still have the issue of too much error handling cluttering normal code.

## Goals
* I'd like to have an easier to use interface for reading/writing binary data.
* I'd like to declare binary IO operations, execute them, and handle a single error at the end.
* I'd like to be able to reuse binary IO operations, and even pass them into more complex pipelines.
* I'd like to be able to declare dynamic behavior, like when the size of the next read is determined by the current field.
* I'd like to declare a read loop based on a read field value, and pass the loop construct to a larger pipeline.
* ~~Struct tag field binding would be fantastic, but reflection is... fraught. I'll see how this goes, and I'll probably take some hints from how the stdlib is handling this.~~
  * There's too much possibility of dynamic logic with a lot of binary payloads, and the number of edge cases for implementing this is more than I want to deal with.
  * I'm pretty happy with the API for mapping definition so far, and I'd rather simplify that than get into reflection with struct field tags. I feel like it's much more understandable (and thus maintainable) code.
