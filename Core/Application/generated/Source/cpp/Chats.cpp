/*
    Chats.cpp
    Generated with 'sql2code' 1.0.0-SNAPSHOT
    https://github.com/red-elf/SQL-to-Code
*/

#include "Chats.h"

std::string Chats::getId() {    
    return this->id;
}

void Chats::setId(std::string value) {
    this->id = value;
}

std::string Chats::getTitle() {    
    return this->title;
}

void Chats::setTitle(std::string value) {
    this->title = value;
}

std::string Chats::getOrganizationId() {    
    return this->organizationId;
}

void Chats::setOrganizationId(std::string value) {
    this->organizationId = value;
}

std::string Chats::getTeamId() {    
    return this->teamId;
}

void Chats::setTeamId(std::string value) {
    this->teamId = value;
}

std::string Chats::getProjectId() {    
    return this->projectId;
}

void Chats::setProjectId(std::string value) {
    this->projectId = value;
}

std::string Chats::getTicketId() {    
    return this->ticketId;
}

void Chats::setTicketId(std::string value) {
    this->ticketId = value;
}

int Chats::getCreated() {    
    return this->created;
}

void Chats::setCreated(int value) {
    this->created = value;
}

int Chats::getModified() {    
    return this->modified;
}

void Chats::setModified(int value) {
    this->modified = value;
}

bool Chats::isDeleted() {    
    return this->deleted;
}

void Chats::setDeleted(bool value) {
    this->deleted = value;
}

