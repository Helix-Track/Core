/*
    TicketTypeProjectMappings.h
    Generated with 'sql2code' 1.0.0-SNAPSHOT
    https://github.com/red-elf/SQL-to-Code
*/

#include "string"

class TicketTypeProjectMappings {

private:
    std::string id;
    std::string ticketTypeId;
    std::string projectId;
    int created;
    int modified;
    bool deleted;

public:
    std::string getId();
    void setId(std::string value);
    std::string getTicketTypeId();
    void setTicketTypeId(std::string value);
    std::string getProjectId();
    void setProjectId(std::string value);
    int getCreated();
    void setCreated(int value);
    int getModified();
    void setModified(int value);
    bool isDeleted();
    void setDeleted(bool value);
};