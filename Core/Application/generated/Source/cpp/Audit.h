/*
    Audit.h
    Generated with 'sql2code' 0.0.1
    https://github.com/red-elf/SQL-to-Code
*/

#include "string"

class Audit {

private:
    std::string id;
    int created;
    std::string entity;
    std::string operation;

public:
    std::string getId();
    void setId(std::string value);
    int getCreated();
    void setCreated(int value);
    std::string getEntity();
    void setEntity(std::string value);
    std::string getOperation();
    void setOperation(std::string value);
};