/*
    TicketRelationships.cpp
    Generated with 'sql2code' 1.0.0-SNAPSHOT
    https://github.com/red-elf/SQL-to-Code
*/

#include "TicketRelationships.h"

std::string TicketRelationships::getId() {    
    return this->id;
}

void TicketRelationships::setId(std::string value) {
    this->id = value;
}

std::string TicketRelationships::getTicketRelationshipTypeId() {    
    return this->ticketRelationshipTypeId;
}

void TicketRelationships::setTicketRelationshipTypeId(std::string value) {
    this->ticketRelationshipTypeId = value;
}

std::string TicketRelationships::getTicketId() {    
    return this->ticketId;
}

void TicketRelationships::setTicketId(std::string value) {
    this->ticketId = value;
}

std::string TicketRelationships::getChildTicketId() {    
    return this->childTicketId;
}

void TicketRelationships::setChildTicketId(std::string value) {
    this->childTicketId = value;
}

int TicketRelationships::getCreated() {    
    return this->created;
}

void TicketRelationships::setCreated(int value) {
    this->created = value;
}

int TicketRelationships::getModified() {    
    return this->modified;
}

void TicketRelationships::setModified(int value) {
    this->modified = value;
}

bool TicketRelationships::isDeleted() {    
    return this->deleted;
}

void TicketRelationships::setDeleted(bool value) {
    this->deleted = value;
}

