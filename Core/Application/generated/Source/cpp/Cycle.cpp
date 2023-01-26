/*
    Cycle.cpp
    Generated with 'sql2code' 0.0.3
    https://github.com/red-elf/SQL-to-Code
*/

#include "Cycle.h"

std::string Cycle::getId() {    
    return this->id;
}

void Cycle::setId(std::string &value) {
    this->id = value;
}

int Cycle::getCreated() {    
    return this->created;
}

void Cycle::setCreated(int &value) {
    this->created = value;
}

int Cycle::getModified() {    
    return this->modified;
}

void Cycle::setModified(int &value) {
    this->modified = value;
}

std::string Cycle::getTitle() {    
    return this->title;
}

void Cycle::setTitle(std::string &value) {
    this->title = value;
}

std::string Cycle::getDescription() {    
    return this->description;
}

void Cycle::setDescription(std::string &value) {
    this->description = value;
}

std::string Cycle::getCycleId() {    
    return this->cycleId;
}

void Cycle::setCycleId(std::string &value) {
    this->cycleId = value;
}

int Cycle::getType() {    
    return this->type;
}

void Cycle::setType(int &value) {
    this->type = value;
}

bool Cycle::isDeleted() {    
    return this->deleted;
}

void Cycle::setDeleted(bool &value) {
    this->deleted = value;
}

