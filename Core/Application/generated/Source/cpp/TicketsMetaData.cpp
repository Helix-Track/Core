/*
    TicketsMetaData.cpp
    Generated with 'sql2code' 0.0.1
    https://github.com/red-elf/SQL-to-Code
*/

#include "TicketsMetaData.h"

std::string TicketsMetaData::getId() {    
    return this->id;
}

void TicketsMetaData::setId(std::string value) {
    this->id = value;
}

std::string TicketsMetaData::getTicketId() {    
    return this->ticketId;
}

void TicketsMetaData::setTicketId(std::string value) {
    this->ticketId = value;
}

std::string TicketsMetaData::getProperty() {    
    return this->property;
}

void TicketsMetaData::setProperty(std::string value) {
    this->property = value;
}

std::string TicketsMetaData::getValue() {    
    return this->value;
}

void TicketsMetaData::setValue(std::string value) {
    this->value = value;
}

int TicketsMetaData::getCreated() {    
    return this->created;
}

void TicketsMetaData::setCreated(int value) {
    this->created = value;
}

int TicketsMetaData::getModified() {    
    return this->modified;
}

void TicketsMetaData::setModified(int value) {
    this->modified = value;
}

bool TicketsMetaData::isDeleted() {    
    return this->deleted;
}

void TicketsMetaData::setDeleted(bool value) {
    this->deleted = value;
}

