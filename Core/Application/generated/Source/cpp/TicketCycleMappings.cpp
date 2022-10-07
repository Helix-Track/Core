/*
    TicketCycleMappings.cpp
    Generated with 'sql2code' 0.0.1
    https://github.com/red-elf/SQL-to-Code
*/

#include "TicketCycleMappings.h"

std::string TicketCycleMappings::getId() {    
    return this->id;
}

void TicketCycleMappings::setId(std::string value) {
    this->id = value;
}

std::string TicketCycleMappings::getTicketId() {    
    return this->ticketId;
}

void TicketCycleMappings::setTicketId(std::string value) {
    this->ticketId = value;
}

std::string TicketCycleMappings::getCycleId() {    
    return this->cycleId;
}

void TicketCycleMappings::setCycleId(std::string value) {
    this->cycleId = value;
}

int TicketCycleMappings::getCreated() {    
    return this->created;
}

void TicketCycleMappings::setCreated(int value) {
    this->created = value;
}

int TicketCycleMappings::getModified() {    
    return this->modified;
}

void TicketCycleMappings::setModified(int value) {
    this->modified = value;
}

bool TicketCycleMappings::isDeleted() {    
    return this->deleted;
}

void TicketCycleMappings::setDeleted(bool value) {
    this->deleted = value;
}

