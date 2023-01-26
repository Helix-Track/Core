/*
    Permission.cpp
    Generated with 'sql2code' 0.0.3
    https://github.com/red-elf/SQL-to-Code
*/

#include "Permission.h"

std::string Permission::getId() {    
    return this->id;
}

void Permission::setId(std::string &value) {
    this->id = value;
}

std::string Permission::getTitle() {    
    return this->title;
}

void Permission::setTitle(std::string &value) {
    this->title = value;
}

std::string Permission::getDescription() {    
    return this->description;
}

void Permission::setDescription(std::string &value) {
    this->description = value;
}

int Permission::getCreated() {    
    return this->created;
}

void Permission::setCreated(int &value) {
    this->created = value;
}

int Permission::getModified() {    
    return this->modified;
}

void Permission::setModified(int &value) {
    this->modified = value;
}

bool Permission::isDeleted() {    
    return this->deleted;
}

void Permission::setDeleted(bool &value) {
    this->deleted = value;
}

