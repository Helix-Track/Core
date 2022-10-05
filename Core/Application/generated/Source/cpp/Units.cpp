/*
    Units.cpp
    Generated with 'sql2code' 1.0.0-SNAPSHOT
    https://github.com/red-elf/SQL-to-Code
*/

#include "Units.h"

std::string Units::getId() {    
    return this->id;
}

void Units::setId(std::string value) {
    this->id = value;
}

std::string Units::getTitle() {    
    return this->title;
}

void Units::setTitle(std::string value) {
    this->title = value;
}

std::string Units::getDescription() {    
    return this->description;
}

void Units::setDescription(std::string value) {
    this->description = value;
}

int Units::getCreated() {    
    return this->created;
}

void Units::setCreated(int value) {
    this->created = value;
}

int Units::getModified() {    
    return this->modified;
}

void Units::setModified(int value) {
    this->modified = value;
}

bool Units::isDeleted() {    
    return this->deleted;
}

void Units::setDeleted(bool value) {
    this->deleted = value;
}

