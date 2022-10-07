/*
    Times.cpp
    Generated with 'sql2code' 0.0.1
    https://github.com/red-elf/SQL-to-Code
*/

#include "Times.h"

std::string Times::getId() {    
    return this->id;
}

void Times::setId(std::string value) {
    this->id = value;
}

int Times::getCreated() {    
    return this->created;
}

void Times::setCreated(int value) {
    this->created = value;
}

int Times::getModified() {    
    return this->modified;
}

void Times::setModified(int value) {
    this->modified = value;
}

int Times::getAmount() {    
    return this->amount;
}

void Times::setAmount(int value) {
    this->amount = value;
}

std::string Times::getUnitId() {    
    return this->unitId;
}

void Times::setUnitId(std::string value) {
    this->unitId = value;
}

std::string Times::getTitle() {    
    return this->title;
}

void Times::setTitle(std::string value) {
    this->title = value;
}

std::string Times::getDescription() {    
    return this->description;
}

void Times::setDescription(std::string value) {
    this->description = value;
}

std::string Times::getTicketId() {    
    return this->ticketId;
}

void Times::setTicketId(std::string value) {
    this->ticketId = value;
}

bool Times::isDeleted() {    
    return this->deleted;
}

void Times::setDeleted(bool value) {
    this->deleted = value;
}

