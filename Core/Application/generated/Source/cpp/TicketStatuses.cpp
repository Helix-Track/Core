/*
    TicketStatuses.cpp
    Generated with 'sql2code' 1.0.0-SNAPSHOT
    https://github.com/red-elf/SQL-to-Code
*/

#include "TicketStatuses.h"

std::string TicketStatuses::getId() {    
    return this->id;
}

void TicketStatuses::setId(std::string value) {
    this->id = value;
}

std::string TicketStatuses::getTitle() {    
    return this->title;
}

void TicketStatuses::setTitle(std::string value) {
    this->title = value;
}

std::string TicketStatuses::getDescription() {    
    return this->description;
}

void TicketStatuses::setDescription(std::string value) {
    this->description = value;
}

int TicketStatuses::getCreated() {    
    return this->created;
}

void TicketStatuses::setCreated(int value) {
    this->created = value;
}

int TicketStatuses::getModified() {    
    return this->modified;
}

void TicketStatuses::setModified(int value) {
    this->modified = value;
}

bool TicketStatuses::isDeleted() {    
    return this->deleted;
}

void TicketStatuses::setDeleted(bool value) {
    this->deleted = value;
}

