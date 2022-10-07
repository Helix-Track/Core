/*
    Audit.cpp
    Generated with 'sql2code' 0.0.1
    https://github.com/red-elf/SQL-to-Code
*/

#include "Audit.h"

std::string Audit::getId() {    
    return this->id;
}

void Audit::setId(std::string value) {
    this->id = value;
}

int Audit::getCreated() {    
    return this->created;
}

void Audit::setCreated(int value) {
    this->created = value;
}

std::string Audit::getEntity() {    
    return this->entity;
}

void Audit::setEntity(std::string value) {
    this->entity = value;
}

std::string Audit::getOperation() {    
    return this->operation;
}

void Audit::setOperation(std::string value) {
    this->operation = value;
}

