/*
    UserOrganizationMapping.h
    Generated with 'sql2code' 0.0.3
    https://github.com/red-elf/SQL-to-Code
*/

#include "string"

class UserOrganizationMapping {

private:
    std::string id;
    std::string userId;
    std::string organizationId;
    int created;
    int modified;
    bool deleted;

public:
    std::string getId();
    void setId(std::string &value);
    std::string getUserId();
    void setUserId(std::string &value);
    std::string getOrganizationId();
    void setOrganizationId(std::string &value);
    int getCreated();
    void setCreated(int &value);
    int getModified();
    void setModified(int &value);
    bool isDeleted();
    void setDeleted(bool &value);
};