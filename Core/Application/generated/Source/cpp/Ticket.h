/*
    Ticket.h
    Generated with 'sql2code' 0.0.1
    https://github.com/red-elf/SQL-to-Code
*/

#include "string"

class Ticket {

private:
    std::string id;
    int ticketNumber;
    int position;
    std::string title;
    std::string description;
    int created;
    int modified;
    std::string ticketTypeId;
    std::string ticketStatusId;
    std::string projectId;
    std::string userId;
    double estimation;
    int storyPoints;
    std::string creator;
    bool deleted;

public:
    std::string getId();
    void setId(std::string value);
    int getTicketNumber();
    void setTicketNumber(int value);
    int getPosition();
    void setPosition(int value);
    std::string getTitle();
    void setTitle(std::string value);
    std::string getDescription();
    void setDescription(std::string value);
    int getCreated();
    void setCreated(int value);
    int getModified();
    void setModified(int value);
    std::string getTicketTypeId();
    void setTicketTypeId(std::string value);
    std::string getTicketStatusId();
    void setTicketStatusId(std::string value);
    std::string getProjectId();
    void setProjectId(std::string value);
    std::string getUserId();
    void setUserId(std::string value);
    double getEstimation();
    void setEstimation(double value);
    int getStoryPoints();
    void setStoryPoints(int value);
    std::string getCreator();
    void setCreator(std::string value);
    bool isDeleted();
    void setDeleted(bool value);
};