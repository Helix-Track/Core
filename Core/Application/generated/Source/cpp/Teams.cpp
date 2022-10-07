/*
    Teams.cpp
    Generated with 'sql2code' 0.0.1
    https://github.com/red-elf/SQL-to-Code
*/

#include "Teams.h"

std::string Teams::getId() {    
    return this->id;
}

void Teams::setId(std::string value) {
    this->id = value;
}

std::string Teams::getTitle() {    
    return this->title;
}

void Teams::setTitle(std::string value) {
    this->title = value;
}

std::string Teams::getDescription() {    
    return this->description;
}

void Teams::setDescription(std::string value) {
    this->description = value;
}

int Teams::getCreated() {    
    return this->created;
}

void Teams::setCreated(int value) {
    this->created = value;
}

int Teams::getModified() {    
    return this->modified;
}

void Teams::setModified(int value) {
    this->modified = value;
}

bool Teams::isDeleted() {    
    return this->deleted;
}

void Teams::setDeleted(bool value) {
    this->deleted = value;
}

