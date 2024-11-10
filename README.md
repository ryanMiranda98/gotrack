# GoTrack
---
GoTrack is an easy-to-use version control system like Git, written in Golang.
---
## Commands available
- version: Prints the current build version.
```console
gotrack version
v1.0.0
```

- init: Initializes an empty gotrack repository. Default is current working directory.
- hash-object: creates a hashed object and stores it in the .gotrack directory.
- cat-file: reads content from the hashed object and displays it in the console.


## How GoTrack works?
GoTrack uses the same concepts as Git as it is essentially a replica.
GoTrack has 4 object types - blob, commit, tag, tree.

##### Blob
Actual user data from files.
For a blob, the file size is calculated in bytes and a SHA-1 hash is generated from the header format below.
The first 2 characters of the hash represent the sub-directory name inside the objects directory and the remaining 38 characters are the filename.
Once the dir/filename is created, the header is compressed with zlib and the resulting (binary) data is written to the file.
```
blob <content_size>\0<content>
```

##### Tree
Tree objects are binary objects. They are used to store files and directory structures.
SHA hash points to either a blob or tree object.
100644 - regular file
100755 - executable file
120000 - symbolic link
040000 - directory
```
tree <size>\0
[mode] [path]\0[SHA-1]
```