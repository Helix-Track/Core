![JWT Compatible](https://jwt.io/img/badge-compatible.svg)
![JIRA alternative for the free world!](Assets/Wide_Black.png)

# HelixTrack Core

Core module for the Helix Track.

## Development

The HelixTrack Core has been developed and tested on [Alt Linux](https://alt-linux.ru/).

## Before you start

TBD: Executing init. scripts.

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

The system scripts and tools located in the `Scripts` directory.

All scripts and tools required for the system to initialize (database, generated code, etc.) are located here.

## Technical documentation

Documentation can be found [here](Documentation).


