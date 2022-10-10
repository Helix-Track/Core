/*
    TicketCycleMapping.cpp
    Generated with 'sql2code' 0.0.2-SNAPSHOT
    https://github.com/red-elf/SQL-to-Code
*/

#include "TicketCycleMapping.h"

std::string TicketCycleMapping::getId() {    
    return this->id;
}

void TicketCycleMapping::setId(std::string &value) {
    this->id = value;
}

std::string TicketCycleMapping::getTicketId() {    
    return this->ticketId;
}

void TicketCycleMapping::setTicketId(std::string &value) {
    this->ticketId = value;
}

std::string TicketCycleMapping::getCycleId() {    
    return this->cycleId;
}

void TicketCycleMapping::setCycleId(std::string &value) {
    this->cycleId = value;
}

int TicketCycleMapping::getCreated() {    
    return this->created;
}

void TicketCycleMapping::setCreated(int &value) {
    this->created = value;
}

int TicketCycleMapping::getModified() {    
    return this->modified;
}

void TicketCycleMapping::setModified(int &value) {
    this->modified = value;
}

bool TicketCycleMapping::isDeleted() {    
    return this->deleted;
}

void TicketCycleMapping::setDeleted(bool &value) {
    this->deleted = value;
}

