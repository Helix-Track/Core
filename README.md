![JWT Compatible](https://jwt.io/img/badge-compatible.svg)
![JIRA alternative for the free world!](Assets/Wide_Black.png)

# HelixTrack Core

The Core module for the Helix Track.

## Development

The HelixTrack Core has been developed and tested on [AltBase Linux distribution](https://www.basealt.ru/).

## Before you start

Clone the project, then, initialize and update the Git submodules.

*Note:* We strongly suggest you to use the `clone` script for this. See next section.

After you have cloned the project execute the `sync` script:

```shell
./sync
```

*Note:* If your Git account does not have the push permissions given by the administrator, instead of the `sync` script do execute `pull_all`:

```shell
./pull_all
```

### Using the `clone` script

To do this automatically execute the following:

```shell
(test -e ./clone || wget "https://raw.githubusercontent.com/red-elf/Project-Toolkit/main/clone?append="$(($(date +%s%N)/1000000)) -O clone) && \
    chmod +x ./clone && ./clone git@github.com:Helix-Track/Core.git ./Core
```

or via one of the mirror repositories:

- [GitFlic](https://gitflic.ru/):

```shell
(test -e ./clone || \
    wget "https://gitflic.ru/project/red-elf/project-toolkit/blob/raw?file=clone&inline=false&append="$(($(date +%s%N)/1000000)) -O clone) && \
    chmod +x ./clone && ./clone git@gitflic.ru:helix-track/core.git ./Core
```

- [Gitee](https://gitee.com/):

```shell
(test -e ./clone || wget "https://gitee.com/Kvetch_Godspeed_b073/Project-Toolkit/raw/main/clone?append="$(($(date +%s%N)/1000000)) -O clone) && \
    chmod +x ./clone && ./clone git@gitee.com:Kvetch_Godspeed_b073/Core.git ./Core
```

*Note:* It is required to execute the script from empty directory where you whish to clone the HelixTrack project.

### Executing the inititialisation scripts

Tbd.

### Opening the project

From the root of the project execute:

```shell
./open
```

*Note:* The `open` command will open the project. Command will perform all the setup work for the project:

- Check if VSCode is available. If it is not, it will download it, and configure it with all mandatory development dependencies,
- Execute all prepare scripts
- Finally, will open the project.

### Testing the project

From the root of the project execute:

```shell
./test
```

It will execute all the [Testable](https://github.com/red-elf/Testable) system components.

## Database

The system database

The `Definition.sqlite` represents the system database. 
It contains all the tables and initial data required for the system to work.

The DDL directory contains all major SQL scripts required to initialize the database.

Convention used for the SQl script is the following:

- The main version scripts:

`Definition.VX.sql` where X represents the version of the database (1, 2, 3, etc).

- Migration scripts:

`Migration.VX.Y.sql` where X represents the version of the database (1, 2, 3, etc) and Y the version of the patch (1, 2, 3, etc).

All SQL scripts are executed by the shell and the `Definition.sqlite` is created as a result.

## Scripts and tools

Tbd.

All scripts and tools required for the system to initialize (database, generated code, etc.)

Tbd.

## Developers documentation

Documentation can be found [here](Documentation).
