/*
    TicketRelationshipType.cpp
    Generated with 'sql2code' 0.0.2-SNAPSHOT
    https://github.com/red-elf/SQL-to-Code
*/

#include "TicketRelationshipType.h"

std::string TicketRelationshipType::getId() {    
    return this->id;
}

void TicketRelationshipType::setId(std::string &value) {
    this->id = value;
}

std::string TicketRelationshipType::getTitle() {    
    return this->title;
}

void TicketRelationshipType::setTitle(std::string &value) {
    this->title = value;
}

std::string TicketRelationshipType::getDescription() {    
    return this->description;
}

void TicketRelationshipType::setDescription(std::string &value) {
    this->description = value;
}

int TicketRelationshipType::getCreated() {    
    return this->created;
}

void TicketRelationshipType::setCreated(int &value) {
    this->created = value;
}

int TicketRelationshipType::getModified() {    
    return this->modified;
}

void TicketRelationshipType::setModified(int &value) {
    this->modified = value;
}

bool TicketRelationshipType::isDeleted() {    
    return this->deleted;
}

void TicketRelationshipType::setDeleted(bool &value) {
    this->deleted = value;
}

