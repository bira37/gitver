# gover

- Simple CLI for semantic versioning

- Controls version following SemVer spec

- Automatically commit and tag version changes to Git (require a clean git repository without any pending changes)

- Possibility to add prefixes to git tags (useful for monorepos to differentiate tags from multiple projects)

Plenty of examples at the end of this document.

## Installation

## Usage

### Flags

```
-d string
      directory: the path for the VERSION file (default "./")
-i string
      increment mode: the increment type. valid inputs: major | minor | patch
-l string
      label: sets a prefix label on the git tag
-new string
      new file: creates a new file in specified directory with version <0.0.0>. Every other flag is ignored
-p string
      prerelease: sets a prerelease suffix
```

### Create a VERSION file

The CLI controls versioning using a file named `VERSION` containing the version string. For the first time only, run:

```sh
gover -new {dir} # to create on specified directory
gover -new ./ # to create on current directory
```

This will create the `VERSION` file. The content of this file will be something like this:

```
0.0.0
```

### How it works

First of all, you must commit every pending change inside the repo before. If git repository is not clean, the command will fail. Starting with a VERSION file in current directory containing `0.0.0`, running `gover -i major -p beta -l myapp` will do the following:

- Change VERSION file content to `1.0.0-beta`
- Create a commit with message `myapp-1.0.0-beta` (shown below)
- Create a tag for the commit with identifier `myapp-1.0.0-beta` (shown below)
```txt
commit 5274c3b2106de103286cad08f5979602effab997 (HEAD -> test, tag: myapp-1.0.0-beta)
Author: Ubiratan Neto <ubiratanneto37@gmail.com>
Date:   Wed Oct 18 20:35:29 2023 -0300
    "myapp-1.0.0-beta"
```

### Examples

- `gover -d . -i major` with file `./VERSION` containg `0.0.0`:
  + VERSION changes to `1.0.0`
  + TAG created as `1.0.0`

- `gover -i major` with file `./VERSION` containg `0.0.0`:
  + VERSION changes to `1.0.0`
  + TAG created as `1.0.0`
  + Note: if -d is ommited, defaults to current directory

- `gover -i patch` with file `./VERSION` containg `1.0.0`:
  + VERSION changes to `1.0.1`
  + TAG created as `1.0.1`

- `gover -i minor -p beta` with file `./VERSION` containg `3.0.2`:
  + VERSION changes to `3.1.0-beta`
  + TAG created as `3.1.0-beta`

- `gover -d services/service-a -l service-a -i minor` with file `./services/service-a/VERSION` containg `1.2.0`:
  + VERSION changes to `1.3.0`
  + TAG created as `service-a-1.3.0`
  + Note: This way, you are allowed to manage versioning across multiple services/packages/libs within a monorepo individually

- `gover -d libs/mylib -l mylib -i major -p beta` with file `./libs/mylib/VERSION` containg `2.5.2`:
  + VERSION changes to `3.0.0-beta`
  + TAG created as `mylib-3.0.0-beta`
