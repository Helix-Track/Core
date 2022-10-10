/*
    Chat.cpp
    Generated with 'sql2code' 0.0.2-SNAPSHOT
    https://github.com/red-elf/SQL-to-Code
*/

#include "Chat.h"

std::string Chat::getId() {    
    return this->id;
}

void Chat::setId(std::string &value) {
    this->id = value;
}

std::string Chat::getTitle() {    
    return this->title;
}

void Chat::setTitle(std::string &value) {
    this->title = value;
}

std::string Chat::getOrganizationId() {    
    return this->organizationId;
}

void Chat::setOrganizationId(std::string &value) {
    this->organizationId = value;
}

std::string Chat::getTeamId() {    
    return this->teamId;
}

void Chat::setTeamId(std::string &value) {
    this->teamId = value;
}

std::string Chat::getProjectId() {    
    return this->projectId;
}

void Chat::setProjectId(std::string &value) {
    this->projectId = value;
}

std::string Chat::getTicketId() {    
    return this->ticketId;
}

void Chat::setTicketId(std::string &value) {
    this->ticketId = value;
}

int Chat::getCreated() {    
    return this->created;
}

void Chat::setCreated(int &value) {
    this->created = value;
}

int Chat::getModified() {    
    return this->modified;
}

void Chat::setModified(int &value) {
    this->modified = value;
}

bool Chat::isDeleted() {    
    return this->deleted;
}

void Chat::setDeleted(bool &value) {
    this->deleted = value;
}

