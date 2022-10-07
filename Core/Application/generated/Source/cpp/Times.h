/*
    Times.h
    Generated with 'sql2code' 0.0.1
    https://github.com/red-elf/SQL-to-Code
*/

#include "string"

class Times {

private:
    std::string id;
    int created;
    int modified;
    int amount;
    std::string unitId;
    std::string title;
    std::string description;
    std::string ticketId;
    bool deleted;

public:
    std::string getId();
    void setId(std::string value);
    int getCreated();
    void setCreated(int value);
    int getModified();
    void setModified(int value);
    int getAmount();
    void setAmount(int value);
    std::string getUnitId();
    void setUnitId(std::string value);
    std::string getTitle();
    void setTitle(std::string value);
    std::string getDescription();
    void setDescription(std::string value);
    std::string getTicketId();
    void setTicketId(std::string value);
    bool isDeleted();
    void setDeleted(bool value);
};