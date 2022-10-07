/*
    TicketProjectMappings.cpp
    Generated with 'sql2code' 0.0.1
    https://github.com/red-elf/SQL-to-Code
*/

#include "TicketProjectMappings.h"

std::string TicketProjectMappings::getId() {    
    return this->id;
}

void TicketProjectMappings::setId(std::string value) {
    this->id = value;
}

std::string TicketProjectMappings::getTicketId() {    
    return this->ticketId;
}

void TicketProjectMappings::setTicketId(std::string value) {
    this->ticketId = value;
}

std::string TicketProjectMappings::getProjectId() {    
    return this->projectId;
}

void TicketProjectMappings::setProjectId(std::string value) {
    this->projectId = value;
}

int TicketProjectMappings::getCreated() {    
    return this->created;
}

void TicketProjectMappings::setCreated(int value) {
    this->created = value;
}

int TicketProjectMappings::getModified() {    
    return this->modified;
}

void TicketProjectMappings::setModified(int value) {
    this->modified = value;
}

bool TicketProjectMappings::isDeleted() {    
    return this->deleted;
}

void TicketProjectMappings::setDeleted(bool value) {
    this->deleted = value;
}

