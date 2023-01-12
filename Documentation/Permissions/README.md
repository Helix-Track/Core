# User permissions

After user is authenticated with success and got the information about the particular HelixTrack Core
instance that it belongs to, permissions information is obtained from that HelixTrack Core instance and returned 
as the part of JWT token's payload.

Each user will have a list of permissions. The following example illustrates regular user with its permissions:

```yaml
{
  "permissions": [
    {
      "permission_id":         "string",
      "permission_context_id": "string"
    }
  ]
}
```

## Permission ID

Permission IDs are connected to the one of the following permissions (with each the proper access level numeric value is associated):

- `READ`  : Allowed reading of the context,                     access level = 1
- `CREATE`: Allowed insertion into the context,                 access level = 2
- `UPDATE`: Allowed modification of the context,                access level = 3
- `DELETE`: Allowed removal of the context,                     access level = 5
- `ALL`   : Allowed to perform all operations on the context    access level = 5

## Permission context ID

Permission context IDs are connected to the one of the following contexts:

- `account`: Access to the accounts
- `account.ACCOUNT_ID`: Access to the account
- `organization`: Access to the organizations (requires access to the account)
- `organization.ORGANIZATION_ID`: Access to the organization (requires access to the account)
- `team`: Access to the teams (requires access to the organization)
- `team.TEAM_ID`: Access to the team (requires access to the organization)
- `project`: Access to the projects (requires access to the organization)
- `project.PROJECT_ID`: Access to the project (requires access to the organization)

*Note:* More permissions contexts to be documented soon.

## Permission contexts hierarchy

- `node`
  - `node.NODE_ID`
    - `account`
      - `account.ACCOUNT_ID`
        - `organization`
          - `organization.ORGANIZATION_ID`
            - `team`
              - `team.TEAM_ID`
            - `project`
              - `project.PROJECT_ID`

*Note:* More permissions context hierarchy to be documented soon.


## How do user permissions work?

For each context where we want to perform certain operation we will verify if that operation is possible to perform by evaluating the following rules:

- Do I have access to the context? If we have access to the context or to a parent context (higher in the hierarchy) the access is granted.
  - No, reject.
  - Yes, lets go to the next check step.

- Do I have propper permission access level? Each operation that we want to execute requires certain level. Let's say that we want to read the content of the context. We need level >= 1. User has the level of 2 (creation granted). That means that it is allowed to read as well.
  - No, reject.
  - Yes, perform the desired operation.