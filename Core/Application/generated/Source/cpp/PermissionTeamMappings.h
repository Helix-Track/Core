/*
    PermissionTeamMappings.h
    Generated with 'sql2code' 1.0.0-SNAPSHOT
    https://github.com/red-elf/SQL-to-Code
*/

#include "string"

class PermissionTeamMappings {

private:
    std::string id;
    std::string permissionId;
    std::string teamId;
    std::string permissionContextId;
    int created;
    int modified;
    bool deleted;

public:
    std::string getId();
    void setId(std::string value);
    std::string getPermissionId();
    void setPermissionId(std::string value);
    std::string getTeamId();
    void setTeamId(std::string value);
    std::string getPermissionContextId();
    void setPermissionContextId(std::string value);
    int getCreated();
    void setCreated(int value);
    int getModified();
    void setModified(int value);
    bool isDeleted();
    void setDeleted(bool value);
};