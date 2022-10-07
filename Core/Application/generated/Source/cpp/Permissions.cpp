/*
    Permissions.cpp
    Generated with 'sql2code' 0.0.1
    https://github.com/red-elf/SQL-to-Code
*/

#include "Permissions.h"

std::string Permissions::getId() {    
    return this->id;
}

void Permissions::setId(std::string value) {
    this->id = value;
}

std::string Permissions::getTitle() {    
    return this->title;
}

void Permissions::setTitle(std::string value) {
    this->title = value;
}

std::string Permissions::getDescription() {    
    return this->description;
}

void Permissions::setDescription(std::string value) {
    this->description = value;
}

int Permissions::getCreated() {    
    return this->created;
}

void Permissions::setCreated(int value) {
    this->created = value;
}

int Permissions::getModified() {    
    return this->modified;
}

void Permissions::setModified(int value) {
    this->modified = value;
}

bool Permissions::isDeleted() {    
    return this->deleted;
}

void Permissions::setDeleted(bool value) {
    this->deleted = value;
}

