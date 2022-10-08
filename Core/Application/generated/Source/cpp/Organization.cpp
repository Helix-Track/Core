/*
    Organization.cpp
    Generated with 'sql2code' 0.0.1
    https://github.com/red-elf/SQL-to-Code
*/

#include "Organization.h"

std::string Organization::getId() {    
    return this->id;
}

void Organization::setId(std::string value) {
    this->id = value;
}

std::string Organization::getTitle() {    
    return this->title;
}

void Organization::setTitle(std::string value) {
    this->title = value;
}

std::string Organization::getDescription() {    
    return this->description;
}

void Organization::setDescription(std::string value) {
    this->description = value;
}

int Organization::getCreated() {    
    return this->created;
}

void Organization::setCreated(int value) {
    this->created = value;
}

int Organization::getModified() {    
    return this->modified;
}

void Organization::setModified(int value) {
    this->modified = value;
}

bool Organization::isDeleted() {    
    return this->deleted;
}

void Organization::setDeleted(bool value) {
    this->deleted = value;
}

