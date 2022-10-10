/*
    TicketTypeProjectMapping.cpp
    Generated with 'sql2code' 0.0.2-SNAPSHOT
    https://github.com/red-elf/SQL-to-Code
*/

#include "TicketTypeProjectMapping.h"

std::string TicketTypeProjectMapping::getId() {    
    return this->id;
}

void TicketTypeProjectMapping::setId(std::string &value) {
    this->id = value;
}

std::string TicketTypeProjectMapping::getTicketTypeId() {    
    return this->ticketTypeId;
}

void TicketTypeProjectMapping::setTicketTypeId(std::string &value) {
    this->ticketTypeId = value;
}

std::string TicketTypeProjectMapping::getProjectId() {    
    return this->projectId;
}

void TicketTypeProjectMapping::setProjectId(std::string &value) {
    this->projectId = value;
}

int TicketTypeProjectMapping::getCreated() {    
    return this->created;
}

void TicketTypeProjectMapping::setCreated(int &value) {
    this->created = value;
}

int TicketTypeProjectMapping::getModified() {    
    return this->modified;
}

void TicketTypeProjectMapping::setModified(int &value) {
    this->modified = value;
}

bool TicketTypeProjectMapping::isDeleted() {    
    return this->deleted;
}

void TicketTypeProjectMapping::setDeleted(bool &value) {
    this->deleted = value;
}

