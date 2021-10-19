# tar-formers
A library and tool to modify tar flows at runtime

```bash
$> tar-formers --help
Copyright (c) 2021 - Daniele Rondina

Tar-formers - A golang tool to control tar flows

Usage:
   [command]

Available Commands:
  completion    generate the autocompletion script for the specified shell
  docker-export Export a docker container files to a specified directory.
  help          Help about any command
  portal        Extract a stdin flow or a tar file to a specified directory.

Flags:
  -c, --config string   Tarformers configuration file
  -d, --debug           Enable debug output.
  -h, --help            help for this command
  -v, --version         version for this command

Use " [command] --help" for more information about a command.
```

## Export docker container and apply filter

```bash
$> tar-formers docker-export <container-id> --to ./tmp
```

***
The `docker export` command at the moment doesn't set the
Uname and Gname attribute of the tarball flow so, you
can't use it with map_entities feature.
***

## Extract tar flow related to a specific rules from stdin

```bash
$> tar -cpf - ./pkg | tar-formers portal --stdin --specs rules.yaml --to ./tmp
```

## Extract tar flow related to a specific rules from file

```bash
$> tar-formers portal --file test.tar --to ./tmp -d --specs rules.yaml
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
# If set to false the files created with be with the running user/group.
same_onwer: false

# Set the access and modification time present on tar header. Default false.
same_chtimes: true

# Using the user/group names present on tar header and resolve it.
# Not yet implemented.
# map_entities: false

```
