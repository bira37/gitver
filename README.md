# gitver

- Simple CLI for semantic versioning

- Controls version following SemVer spec

- Automatically push tag versions to Git (require a clean git repository without any pending changes)

- Possibility to add prefixes to git tags (useful for monorepos to differentiate tags from multiple projects)

Plenty of examples at the end of this document.

## Installation

```sh
go install github.com/bira37/gitver
```

## Usage

### Flags

```
-config string
      config: location of the config file. If not provided, try to find in current dir. Config file has priority over directory flag (default "./gitver.json")
-d string
      directory: the path for the VERSION file (default "./")
-i string
      increment mode: the increment type. valid inputs: major | minor | patch
-l string
      label: sets a prefix label only on the git tag (useful for monorepos to differentiate tags from multiple projects)
-p string
      prerelease: sets a prerelease suffix
-r    release: removes prerelease suffix. Overrides prerelease option
```

### How it works

First of all, you must commit every pending change inside the repo before. If git repository is not clean, the command will fail. You need to have at least one tag created in your git repo (such as `0.0.0`). For example, running `gitver -i major -p beta -l myapp` will do the following:

- Create a tag for the latest commit with identifier `myapp-1.0.0-beta` (shown below)
```txt
commit 5274c3b2106de103286cad08f5979602effab997 (HEAD -> test, tag: myapp-1.0.0-beta)
Author: Ubiratan Neto <ubiratanneto37@gmail.com>
Date:   Wed Oct 18 20:35:29 2023 -0300
    "my last commit"
```

### Examples

- `gitver -d . -i major` with latest version as `0.0.0`:
  + VERSION changes to `1.0.0`
  + TAG created as `1.0.0`

- `gitver -i major` with latest version as `0.0.0`:
  + VERSION changes to `1.0.0`
  + TAG created as `1.0.0`
  + Note: if -d is ommited, defaults to current directory

- `gitver -i patch` with latest version as `1.0.0`:
  + VERSION changes to `1.0.1`
  + TAG created as `1.0.1`

- `gitver -i minor -p beta` with latest version as `3.0.2`:
  + VERSION changes to `3.1.0-beta`
  + TAG created as `3.1.0-beta`

- `gitver -d services/service-a -l service-a -i minor` with latest tag for `service-a` as `service-a-1.2.0`:
  + VERSION changes to `1.3.0`
  + TAG created as `service-a-1.3.0`
  + Note: This way, you are allowed to manage versioning across multiple services/packages/libs within a monorepo individually

- `gitver -d libs/mylib -l mylib -i major -p beta` with latest tag for `mylib` as `mylib-2.5.2`:
  + VERSION changes to `3.0.0-beta`
  + TAG created as `mylib-3.0.0-beta`

### Configs

There are some other changes you can make to the tag generation flow using a `gitver.json` file, the complete JSON is shown below:

```json
{
  "allowedLabels": ["mylib"],
  "allowedBranches": ["main"]
}

```

- `allowedLabels` only allows labels specified on config file. Allow everything if not passed
- `allowedBranches` only allows tag generation on specified branches. Allow every branch if not passed