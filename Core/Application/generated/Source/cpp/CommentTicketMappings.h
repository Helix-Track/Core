/*
    CommentTicketMappings.h
    Generated with 'sql2code' 1.0.0-SNAPSHOT
    https://github.com/red-elf/SQL-to-Code
*/

#include "string"

class CommentTicketMappings {

private:
    std::string id;
    std::string commentId;
    std::string ticketId;
    int created;
    int modified;
    bool deleted;

public:
    std::string getId();
    void setId(std::string value);
    std::string getCommentId();
    void setCommentId(std::string value);
    std::string getTicketId();
    void setTicketId(std::string value);
    int getCreated();
    void setCreated(int value);
    int getModified();
    void setModified(int value);
    bool isDeleted();
    void setDeleted(bool value);
};