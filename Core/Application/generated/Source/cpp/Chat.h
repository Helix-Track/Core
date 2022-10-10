/*
    Chat.h
    Generated with 'sql2code' 0.0.2-SNAPSHOT
    https://github.com/red-elf/SQL-to-Code
*/

#include "string"

class Chat {

private:
    std::string id;
    std::string title;
    std::string organizationId;
    std::string teamId;
    std::string projectId;
    std::string ticketId;
    int created;
    int modified;
    bool deleted;

public:
    std::string getId();
    void setId(std::string &value);
    std::string getTitle();
    void setTitle(std::string &value);
    std::string getOrganizationId();
    void setOrganizationId(std::string &value);
    std::string getTeamId();
    void setTeamId(std::string &value);
    std::string getProjectId();
    void setProjectId(std::string &value);
    std::string getTicketId();
    void setTicketId(std::string &value);
    int getCreated();
    void setCreated(int &value);
    int getModified();
    void setModified(int &value);
    bool isDeleted();
    void setDeleted(bool &value);
};