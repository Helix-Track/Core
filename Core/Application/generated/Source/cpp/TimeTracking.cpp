/*
    TimeTracking.cpp
    Generated with 'sql2code' 0.0.2-SNAPSHOT
    https://github.com/red-elf/SQL-to-Code
*/

#include "TimeTracking.h"

std::string TimeTracking::getId() {    
    return this->id;
}

void TimeTracking::setId(std::string &value) {
    this->id = value;
}

int TimeTracking::getCreated() {    
    return this->created;
}

void TimeTracking::setCreated(int &value) {
    this->created = value;
}

int TimeTracking::getModified() {    
    return this->modified;
}

void TimeTracking::setModified(int &value) {
    this->modified = value;
}

int TimeTracking::getAmount() {    
    return this->amount;
}

void TimeTracking::setAmount(int &value) {
    this->amount = value;
}

std::string TimeTracking::getUnitId() {    
    return this->unitId;
}

void TimeTracking::setUnitId(std::string &value) {
    this->unitId = value;
}

std::string TimeTracking::getTitle() {    
    return this->title;
}

void TimeTracking::setTitle(std::string &value) {
    this->title = value;
}

std::string TimeTracking::getDescription() {    
    return this->description;
}

void TimeTracking::setDescription(std::string &value) {
    this->description = value;
}

std::string TimeTracking::getTicketId() {    
    return this->ticketId;
}

void TimeTracking::setTicketId(std::string &value) {
    this->ticketId = value;
}

bool TimeTracking::isDeleted() {    
    return this->deleted;
}

void TimeTracking::setDeleted(bool &value) {
    this->deleted = value;
}

