/*
    TicketProjectMapping.cpp
    Generated with 'sql2code' 0.0.1
    https://github.com/red-elf/SQL-to-Code
*/

#include "TicketProjectMapping.h"

std::string TicketProjectMapping::getId() {    
    return this->id;
}

void TicketProjectMapping::setId(std::string value) {
    this->id = value;
}

std::string TicketProjectMapping::getTicketId() {    
    return this->ticketId;
}

void TicketProjectMapping::setTicketId(std::string value) {
    this->ticketId = value;
}

std::string TicketProjectMapping::getProjectId() {    
    return this->projectId;
}

void TicketProjectMapping::setProjectId(std::string value) {
    this->projectId = value;
}

int TicketProjectMapping::getCreated() {    
    return this->created;
}

void TicketProjectMapping::setCreated(int value) {
    this->created = value;
}

int TicketProjectMapping::getModified() {    
    return this->modified;
}

void TicketProjectMapping::setModified(int value) {
    this->modified = value;
}

bool TicketProjectMapping::isDeleted() {    
    return this->deleted;
}

void TicketProjectMapping::setDeleted(bool value) {
    this->deleted = value;
}

