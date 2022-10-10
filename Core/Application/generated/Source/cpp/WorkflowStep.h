/*
    WorkflowStep.h
    Generated with 'sql2code' 0.0.2-SNAPSHOT
    https://github.com/red-elf/SQL-to-Code
*/

#include "string"

class WorkflowStep {

private:
    std::string id;
    std::string title;
    std::string description;
    std::string workflowId;
    std::string workflowStepId;
    std::string ticketStatusId;
    int created;
    int modified;
    bool deleted;

public:
    std::string getId();
    void setId(std::string &value);
    std::string getTitle();
    void setTitle(std::string &value);
    std::string getDescription();
    void setDescription(std::string &value);
    std::string getWorkflowId();
    void setWorkflowId(std::string &value);
    std::string getWorkflowStepId();
    void setWorkflowStepId(std::string &value);
    std::string getTicketStatusId();
    void setTicketStatusId(std::string &value);
    int getCreated();
    void setCreated(int &value);
    int getModified();
    void setModified(int &value);
    bool isDeleted();
    void setDeleted(bool &value);
};