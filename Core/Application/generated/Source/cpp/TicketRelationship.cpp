/*
    TicketRelationship.cpp
    Generated with 'sql2code' 0.0.2-SNAPSHOT
    https://github.com/red-elf/SQL-to-Code
*/

#include "TicketRelationship.h"

std::string TicketRelationship::getId() {    
    return this->id;
}

void TicketRelationship::setId(std::string &value) {
    this->id = value;
}

std::string TicketRelationship::getTicketRelationshipTypeId() {    
    return this->ticketRelationshipTypeId;
}

void TicketRelationship::setTicketRelationshipTypeId(std::string &value) {
    this->ticketRelationshipTypeId = value;
}

std::string TicketRelationship::getTicketId() {    
    return this->ticketId;
}

void TicketRelationship::setTicketId(std::string &value) {
    this->ticketId = value;
}

std::string TicketRelationship::getChildTicketId() {    
    return this->childTicketId;
}

void TicketRelationship::setChildTicketId(std::string &value) {
    this->childTicketId = value;
}

int TicketRelationship::getCreated() {    
    return this->created;
}

void TicketRelationship::setCreated(int &value) {
    this->created = value;
}

int TicketRelationship::getModified() {    
    return this->modified;
}

void TicketRelationship::setModified(int &value) {
    this->modified = value;
}

bool TicketRelationship::isDeleted() {    
    return this->deleted;
}

void TicketRelationship::setDeleted(bool &value) {
    this->deleted = value;
}

