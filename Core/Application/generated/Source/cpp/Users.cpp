/*
    Users.cpp
    Generated with 'sql2code' 0.0.1
    https://github.com/red-elf/SQL-to-Code
*/

#include "Users.h"

std::string Users::getId() {    
    return this->id;
}

void Users::setId(std::string value) {
    this->id = value;
}

int Users::getCreated() {    
    return this->created;
}

void Users::setCreated(int value) {
    this->created = value;
}

int Users::getModified() {    
    return this->modified;
}

void Users::setModified(int value) {
    this->modified = value;
}

bool Users::isDeleted() {    
    return this->deleted;
}

void Users::setDeleted(bool value) {
    this->deleted = value;
}

