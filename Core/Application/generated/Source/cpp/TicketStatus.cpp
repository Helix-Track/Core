/*
    TicketStatus.cpp
    Generated with 'sql2code' 0.0.1
    https://github.com/red-elf/SQL-to-Code
*/

#include "TicketStatus.h"

std::string TicketStatus::getId() {    
    return this->id;
}

void TicketStatus::setId(std::string value) {
    this->id = value;
}

std::string TicketStatus::getTitle() {    
    return this->title;
}

void TicketStatus::setTitle(std::string value) {
    this->title = value;
}

std::string TicketStatus::getDescription() {    
    return this->description;
}

void TicketStatus::setDescription(std::string value) {
    this->description = value;
}

int TicketStatus::getCreated() {    
    return this->created;
}

void TicketStatus::setCreated(int value) {
    this->created = value;
}

int TicketStatus::getModified() {    
    return this->modified;
}

void TicketStatus::setModified(int value) {
    this->modified = value;
}

bool TicketStatus::isDeleted() {    
    return this->deleted;
}

void TicketStatus::setDeleted(bool value) {
    this->deleted = value;
}

