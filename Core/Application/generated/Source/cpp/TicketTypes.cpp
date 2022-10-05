/*
    TicketTypes.cpp
    Generated with 'sql2code' 1.0.0-SNAPSHOT
    https://github.com/red-elf/SQL-to-Code
*/

#include "TicketTypes.h"

std::string TicketTypes::getId() {    
    return this->id;
}

void TicketTypes::setId(std::string value) {
    this->id = value;
}

std::string TicketTypes::getTitle() {    
    return this->title;
}

void TicketTypes::setTitle(std::string value) {
    this->title = value;
}

std::string TicketTypes::getDescription() {    
    return this->description;
}

void TicketTypes::setDescription(std::string value) {
    this->description = value;
}

int TicketTypes::getCreated() {    
    return this->created;
}

void TicketTypes::setCreated(int value) {
    this->created = value;
}

int TicketTypes::getModified() {    
    return this->modified;
}

void TicketTypes::setModified(int value) {
    this->modified = value;
}

bool TicketTypes::isDeleted() {    
    return this->deleted;
}

void TicketTypes::setDeleted(bool value) {
    this->deleted = value;
}

