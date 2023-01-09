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

Permission IDs are connected to the one of the following permissions:

- `CREATE`: Allowed insertion into the context
- `UPDATE`: Allowed modification of the context
- `DELETE`: Allowed removal of the context
- `ALL`   : Allowed to perform all operations on the context

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

## How do user permissions work?

Tbd.