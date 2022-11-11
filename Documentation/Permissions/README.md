# User roles

User can belong to the one (or more) of the predefined `roles`:

## Role `root` 

Root access to everything on the particular HelixTrack Core instance.

## Role `admin`: 

Admin access to everything under the particular account.

## Role `user`

Regularly registered user account. By the default it does not have any access
level. Access level can be gained by:

- Registering `account` by the user and taking its ownership (becoming `admin`)
- Receiving various permissions from the `admin` or the `root` users
- Transferring the ownership from the `admin` or the `root` users.

# User permissions

Each user will have a list of permissions. Each permission has the structure.

The following example illustrates regular user with its permissions:

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

## Permission IDs

- `CREATE`: Allowed insertion into the permission context
- `UPDATE`: Allowed modification of the permission context
- `DELETE`: Allowed removal of the permission context

## Permission contexts

- `organization`: Access to the organizations
- `organization.ORGANIZATION_ID`: Access to the organization
- `project`: Access to the projects (requires access to the organization)
- `project.PROJECT_ID`: Access to the project (requires access to the organization)

*Note:* More permissions contexts to be documented soon.

## How do user permissions work?

Tbd.