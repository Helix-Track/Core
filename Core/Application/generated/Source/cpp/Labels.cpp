/*
    Labels.cpp
    Generated with 'sql2code' 1.0.0-SNAPSHOT
    https://github.com/red-elf/SQL-to-Code
*/

#include "Labels.h"

std::string Labels::getId() {    
    return this->id;
}

void Labels::setId(std::string value) {
    this->id = value;
}

std::string Labels::getTitle() {    
    return this->title;
}

void Labels::setTitle(std::string value) {
    this->title = value;
}

std::string Labels::getDescription() {    
    return this->description;
}

void Labels::setDescription(std::string value) {
    this->description = value;
}

int Labels::getCreated() {    
    return this->created;
}

void Labels::setCreated(int value) {
    this->created = value;
}

int Labels::getModified() {    
    return this->modified;
}

void Labels::setModified(int value) {
    this->modified = value;
}

bool Labels::isDeleted() {    
    return this->deleted;
}

void Labels::setDeleted(bool value) {
    this->deleted = value;
}

