![JIRA alternative for the free world!](Assets/Wide_Black.png)

# Core

The Core module for the Helix Track.

The following directory structure represents the layout of the project:

- Assets

Contains all the static assets required for the project (images, documents and other static items).

- Core

Contains the Open-source part of the project.

- Propriatery

Contains the propriatery part of the project.

The Core and Propriatery directories have the following structure.

- Database
- Scripts
- Source

## Database

The system database

The `Definition.sqlite` represents the system database. 
It contains all the tables and initial data required for the system to work.

The DDL directory contains all major SQL scripts required to initialize the database.

Convention used for the SQl script is the following:

- The main version scripts:

`Definition.VX.sql` where X represnts the version of the database (1, 2, 3, etc).

- Migration scripts:

`Migration.VX.Y.sql` where X represnts the version of the database (1, 2, 3, etc) and Y the version of the patch (1, 2, 3, etc).

All SQL scripts are executed by the shell and the `Definition.sqlite` is created as a result.

## Scripts

The system scripts

All scripts required for the system to intialize (database, generated code, etc.) are located here.

Scripts are divided as:

- Standalone scripts
- Utility scripts.

The naming convention used for the naming is the following:

- `Main.Purpose.sh`, for the standalone scripts, where the 'Purpose' is the name of script's main functionality
- `Util.Purpose.sh`, for the utility scripts, where the 'Purpose' is the name of script's utility functionality.

## Source

The application layer source code.

All the application source code goies here.

Source code is divided by the applications.

Each application has the separate Git repository (git submodule).

The naming convention used for the naming is the following:

- `Helix XXX`, where the 'XXX' part represents the name of the application.


