# watchrun

This is a small program which watches a file or a directory for changes and executes
a command when this happens.

## Usage

```sh
watchrun ~/watched.txt echo The File Changed
```

Alternatively you can watch a directory:

```sh
watchrun $HOME echo A file in my home has changed 
```

Directories are *not* watched recursively. The tool will execute the command only when
the file has been changed or created.


## Install

### Released Binary

You can go to the [releases page](https://github.com/ironsmile/watchrun/releases) and choose the binary for your OS.

### From Source

This is Go program using go modules so you can just

```sh
go get github.com/ironsmile/watchrun
```
