/*
    TicketType.cpp
    Generated with 'sql2code' 0.0.2-SNAPSHOT
    https://github.com/red-elf/SQL-to-Code
*/

#include "TicketType.h"

std::string TicketType::getId() {    
    return this->id;
}

void TicketType::setId(std::string &value) {
    this->id = value;
}

std::string TicketType::getTitle() {    
    return this->title;
}

void TicketType::setTitle(std::string &value) {
    this->title = value;
}

std::string TicketType::getDescription() {    
    return this->description;
}

void TicketType::setDescription(std::string &value) {
    this->description = value;
}

int TicketType::getCreated() {    
    return this->created;
}

void TicketType::setCreated(int &value) {
    this->created = value;
}

int TicketType::getModified() {    
    return this->modified;
}

void TicketType::setModified(int &value) {
    this->modified = value;
}

bool TicketType::isDeleted() {    
    return this->deleted;
}

void TicketType::setDeleted(bool &value) {
    this->deleted = value;
}

