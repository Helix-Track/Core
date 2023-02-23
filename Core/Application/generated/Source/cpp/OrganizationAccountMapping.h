/*
    OrganizationAccountMapping.h
    Generated with 'sql2code' 0.0.3
    https://github.com/red-elf/SQL-to-Code
*/

#include "string"

class OrganizationAccountMapping {

private:
    std::string id;
    std::string organizationId;
    std::string accountId;
    int created;
    int modified;
    bool deleted;

public:
    std::string getId();
    void setId(std::string &value);
    std::string getOrganizationId();
    void setOrganizationId(std::string &value);
    std::string getAccountId();
    void setAccountId(std::string &value);
    int getCreated();
    void setCreated(int &value);
    int getModified();
    void setModified(int &value);
    bool isDeleted();
    void setDeleted(bool &value);
};