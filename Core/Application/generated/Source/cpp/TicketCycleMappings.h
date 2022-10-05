/*
    TicketCycleMappings.h
    Generated with 'sql2code' 1.0.0-SNAPSHOT
    https://github.com/red-elf/SQL-to-Code
*/

#include "string"

class TicketCycleMappings {

private:
    std::string id;
    std::string ticketId;
    std::string cycleId;
    int created;
    int modified;
    bool deleted;

public:
    std::string getId();
    void setId(std::string value);
    std::string getTicketId();
    void setTicketId(std::string value);
    std::string getCycleId();
    void setCycleId(std::string value);
    int getCreated();
    void setCreated(int value);
    int getModified();
    void setModified(int value);
    bool isDeleted();
    void setDeleted(bool value);
};