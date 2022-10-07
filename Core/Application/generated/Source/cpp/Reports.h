/*
    Reports.h
    Generated with 'sql2code' 0.0.1
    https://github.com/red-elf/SQL-to-Code
*/

#include "string"

class Reports {

private:
    std::string id;
    int created;
    int modified;
    std::string title;
    std::string description;
    bool deleted;

public:
    std::string getId();
    void setId(std::string value);
    int getCreated();
    void setCreated(int value);
    int getModified();
    void setModified(int value);
    std::string getTitle();
    void setTitle(std::string value);
    std::string getDescription();
    void setDescription(std::string value);
    bool isDeleted();
    void setDeleted(bool value);
};