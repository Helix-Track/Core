/*
    TicketMetaData.cpp
    Generated with 'sql2code' 0.0.2-SNAPSHOT
    https://github.com/red-elf/SQL-to-Code
*/

#include "TicketMetaData.h"

std::string TicketMetaData::getId() {    
    return this->id;
}

void TicketMetaData::setId(std::string &value) {
    this->id = value;
}

std::string TicketMetaData::getTicketId() {    
    return this->ticketId;
}

void TicketMetaData::setTicketId(std::string &value) {
    this->ticketId = value;
}

std::string TicketMetaData::getProperty() {    
    return this->property;
}

void TicketMetaData::setProperty(std::string &value) {
    this->property = value;
}

std::string TicketMetaData::getValue() {    
    return this->value;
}

void TicketMetaData::setValue(std::string &value) {
    this->value = value;
}

int TicketMetaData::getCreated() {    
    return this->created;
}

void TicketMetaData::setCreated(int &value) {
    this->created = value;
}

int TicketMetaData::getModified() {    
    return this->modified;
}

void TicketMetaData::setModified(int &value) {
    this->modified = value;
}

bool TicketMetaData::isDeleted() {    
    return this->deleted;
}

void TicketMetaData::setDeleted(bool &value) {
    this->deleted = value;
}

