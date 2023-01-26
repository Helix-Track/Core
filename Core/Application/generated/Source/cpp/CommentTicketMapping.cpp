/*
    CommentTicketMapping.cpp
    Generated with 'sql2code' 0.0.3
    https://github.com/red-elf/SQL-to-Code
*/

#include "CommentTicketMapping.h"

std::string CommentTicketMapping::getId() {    
    return this->id;
}

void CommentTicketMapping::setId(std::string &value) {
    this->id = value;
}

std::string CommentTicketMapping::getCommentId() {    
    return this->commentId;
}

void CommentTicketMapping::setCommentId(std::string &value) {
    this->commentId = value;
}

std::string CommentTicketMapping::getTicketId() {    
    return this->ticketId;
}

void CommentTicketMapping::setTicketId(std::string &value) {
    this->ticketId = value;
}

int CommentTicketMapping::getCreated() {    
    return this->created;
}

void CommentTicketMapping::setCreated(int &value) {
    this->created = value;
}

int CommentTicketMapping::getModified() {    
    return this->modified;
}

void CommentTicketMapping::setModified(int &value) {
    this->modified = value;
}

bool CommentTicketMapping::isDeleted() {    
    return this->deleted;
}

void CommentTicketMapping::setDeleted(bool &value) {
    this->deleted = value;
}

