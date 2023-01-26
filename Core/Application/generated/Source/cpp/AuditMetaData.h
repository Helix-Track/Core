/*
    AuditMetaData.h
    Generated with 'sql2code' 0.0.3
    https://github.com/red-elf/SQL-to-Code
*/

#include "string"

class AuditMetaData {

private:
    std::string id;
    std::string auditId;
    std::string property;
    std::string value;
    int created;
    int modified;

public:
    std::string getId();
    void setId(std::string &value);
    std::string getAuditId();
    void setAuditId(std::string &value);
    std::string getProperty();
    void setProperty(std::string &value);
    std::string getValue();
    void setValue(std::string &value);
    int getCreated();
    void setCreated(int &value);
    int getModified();
    void setModified(int &value);
};