/*
    Cycle.h
    Generated with 'sql2code' 0.0.3
    https://github.com/red-elf/SQL-to-Code
*/

#include "string"

class Cycle {

private:
    std::string id;
    int created;
    int modified;
    std::string title;
    std::string description;
    std::string cycleId;
    int type;
    bool deleted;

public:
    std::string getId();
    void setId(std::string &value);
    int getCreated();
    void setCreated(int &value);
    int getModified();
    void setModified(int &value);
    std::string getTitle();
    void setTitle(std::string &value);
    std::string getDescription();
    void setDescription(std::string &value);
    std::string getCycleId();
    void setCycleId(std::string &value);
    int getType();
    void setType(int &value);
    bool isDeleted();
    void setDeleted(bool &value);
};