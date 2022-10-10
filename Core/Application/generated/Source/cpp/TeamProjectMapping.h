/*
    TeamProjectMapping.h
    Generated with 'sql2code' 0.0.2
    https://github.com/red-elf/SQL-to-Code
*/

#include "string"

class TeamProjectMapping {

private:
    std::string id;
    std::string teamId;
    std::string projectId;
    int created;
    int modified;
    bool deleted;

public:
    std::string getId();
    void setId(std::string &value);
    std::string getTeamId();
    void setTeamId(std::string &value);
    std::string getProjectId();
    void setProjectId(std::string &value);
    int getCreated();
    void setCreated(int &value);
    int getModified();
    void setModified(int &value);
    bool isDeleted();
    void setDeleted(bool &value);
};