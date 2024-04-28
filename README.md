# tar-formers

[![Go Report Card](https://goreportcard.com/badge/github.com/geaaru/tar-formers)](https://goreportcard.com/report/github.com/geaaru/tar-formers)
[![Build on push](https://github.com/geaaru/tar-formers/actions/workflows/push.yml/badge.svg)](https://github.com/geaaru/tar-formers/actions/workflows/push.yml)

A library and tool to modify tar flows/streams at runtime.

The tool `tar-formers` was born to be a helpful tool for testing the
library tar-formers that is mainly used by the [luet](https://github.com/geaaru/luet) the
[Macaroni OS](https://github.com/macaroni-os/) PMS.
But could be used as static binary for archiving the directories, and files in different
compressions as an alternative to `tar` binary and apply renames, and filters at runtime.

```bash
Copyright (c) 2021-2024 - Daniele Rondina

Tar-formers - A golang tool to control tar flows/streams

Usage:
   [command]

Available Commands:
  archive       Archive one or more directories to a tarball.
  bridge        Extract a stdin flow or an input tarball and bridge it to tar output stream or file.
  completion    Generate the autocompletion script for the specified shell
  docker-cp     Copy files from a docker container path to a specified directory or to a file.
  docker-export Export the files a docker container to a specified directory or to a file.
  docker-import Create a docker image from a directory or a tarball.
  help          Help about any command
  portal        Extract a stdin flow or a tar file to a specified directory.

Flags:
  -c, --config string   Tarformers configuration file
  -d, --debug           Enable debug output.
  -h, --help            help for this command
  -v, --version         version for this command

Use " [command] --help" for more information about a command.
```

## Export docker container to a directory and apply filter

```bash
$> tar-formers docker-export <container-id> --todir ./tmp
```

***
The `docker export` command at the moment doesn't set the
Uname and Gname attribute of the tarball flow so, you
can't use it with map_entities feature.
***

## Export docker container to a tarball and apply filter

The type of the compression is automatically detected by the
extension of the filename.

```bash
$> tar-formers docker-export <container-id> --to /mycontaincer.tar.gz --specs specs.yml
```

The option `--to` accepts the `-` for write flow to stdout.

## Copy files from a docker container to a directory and apply filter

```bash
$> tar-formers docker-cp <container-id> <container-src-path> --todir ./tmp --specs specs.yml
```

## Copy files from a docker container to a tarball and apply filter

```bash
$> tar-formers docker-cp <container-id> <container-src-path> --to /myfiles.tar.gz ./tmp --specs specs.yml
```


## Create a docker image from with the content of a directory filtered

```bash
$> tar-formers docker-import geaaru/tar-formers:latest --dir ./tmp --platform amd64 \
    -m "My image" --change 'ENTRYPOINT ["/bin/sh"]' --specs specs.yml
```

## Create a docker image from a tarball filtered

```bash
$> tar-formers di geaaru/tar-formers:latest --file container.tar --platform amd64 \
    -m "My image" --change 'ENTRYPOINT ["/bin/sh"]' --specs specs.yml
```

NOTE: The supported files for the option `--file` are: gzip|gz,zstd,xz,bzip2|bz2,tar

## Extract tar flow related to a specific rules from stdin

```bash
$> tar -cpf - ./pkg | tar-formers portal --stdin --specs rules.yaml --to ./tmp
```

## Extract tar flow related to a specific rules from tar file compressed in gzip

```bash
$> tar-formers portal --file test.tar.gz --to ./tmp -d --specs rules.yaml
```

### Rules YAML file

`tar-formers` takes a rules YAML file in this format:

```yaml
# Author: geaaru@sabayonlinux.org
# tar-formers example specs file.

# Define the list of path prefix that are accepted
# If the tar entity doesn't match with the defined
# prefix will be ignored.
# An empty list means accept all.
match_prefix:
#- "/etc/"
# NOTE: for tarball with relative path you need
# consider to add an additional slash at begin.
# - "/./pkg/"

# Define the list of regex used to match the paths
# to ignore. This check is done after the match prefix
# rules.
ignore_regexes:
# - "^/var/pkg"
#  - "^/./pkg/specs"


# Define the list of files to ignore.
ignore_files:
  - "/.dockerenv"

# Define the list of files to rename
rename:
  - source: "/etc/resolv.conf"
    dest: "/etc/resolv.conf.example"

# Define a list of uids to remap. The uid is a uint32 number.
# Not yet implemented.
#remap_uids:
#  100: 101

# Define a list of gids to remap. The uid is a uint32 number.
# Not yet implemented.
#remap_gids:
#  1000: 1001

# Set the same owner present on tarfile. Default true.
same_onwer: false

# Set the access and modification time present on tar header. Default false.
same_chtimes: true

# Using the user/group names present on tar header and resolve it.
# Not yet implemented.
# map_entities: false

# Warning on create hardlink and sym
broken_links_fatal: false
```

## Golang API

Hereinafter, an example about using `tar-formers` API:

```go

  // Initialize default tarformers instance
  // to use the config object used by the library.
  cfg := tarf_specs.NewConfig(c.Viper)
  cfg.GetLogging().Level = "warning"

  t := tarf.NewTarFormersWithLog(cfg, true)
  tarf.SetDefaultTarFormers(t)

  // Untar file
  in, err := os.Open(srcTarfile)
  if err != nil {
    return err
  }
  defer in.Close()

  spec := tarf_specs.NewSpecFile()
  spec.SameOwner = true
  spec.EnableMutex = true
  spec.OverwritePerms = true
  spec.IgnoreFiles = []string{
    "/dev",
    "/.dockerenv",
  }

  tarformers := tarf.NewTarFormers(tarf.GetOptimusPrime().Config)
  tarformers.SetReader(in)

  if modifier != nil && len(protectedFiles) > 0 {
    tarformers.SetFileHandler(modifier)

      spec.TriggeredFiles = protectedFiles
  }

  return tarformers.RunTask(spec, dst)
```
