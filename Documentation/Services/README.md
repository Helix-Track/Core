# HelixTrack Services

The following list represents supported HelixTrack Services:

- Core                  (mandatory, main,       opensource                   )
- Authentication        (mandatory, main,       propriatery, provides the API)
- Permissions Engine    (mandatory, main,       propriatery, provides the API)
- Lokalisation          (optional,  extension,  propriatery, provides the API)
- Times                 (optional,  extension                                )
- Documents             (optional,  extension                                )
- Chats                 (optional,  extension                                )

## Services project structure

The services listed above fit in the following project structure:

- `Core`, Main project directory
    - `Run`, Run scripts
    - `Core`, HelixTrack Core
    - `Assets`, Various assets
    - `Documentation`, Project documentation
    - `Propriatery`, The propriatery implementations
    - `Services`, Services directory
        - `Authentication`, Authentication extension
            - `API`, Open-source API
            - `Implementation`, The propriatery implementation
        - `Permissions`, Permissions engine
            - `API`, Open-source API
            - `Implementation`, The propriatery implementation
        - `Lokalisation`, Lokalisation extension
            - `API`, Open-source API
            - `Implementation`, The propriatery implementation
    - `Extensions`, Extensions directory
        - `Times`, Times extension
            - `API`, Open-source API
            - `Implementation`, The propriatery implementation
        - `Documents`, Times extension
            - `API`, Open-source API
            - `Implementation`, The propriatery implementation
        - `Chats`, Times extension
            - `API`, Open-source API
            - `Implementation`, The propriatery implementation
