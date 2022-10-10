/*
    TicketMetaData.h
    Generated with 'sql2code' 0.0.2-SNAPSHOT
    https://github.com/red-elf/SQL-to-Code
*/

#include "string"

class TicketMetaData {

private:
    std::string id;
    std::string ticketId;
    std::string property;
    std::string value;
    int created;
    int modified;
    bool deleted;

public:
    std::string getId();
    void setId(std::string &value);
    std::string getTicketId();
    void setTicketId(std::string &value);
    std::string getProperty();
    void setProperty(std::string &value);
    std::string getValue();
    void setValue(std::string &value);
    int getCreated();
    void setCreated(int &value);
    int getModified();
    void setModified(int &value);
    bool isDeleted();
    void setDeleted(bool &value);
};