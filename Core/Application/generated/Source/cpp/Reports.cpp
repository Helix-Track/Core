/*
    Reports.cpp
    Generated with 'sql2code' 0.0.1
    https://github.com/red-elf/SQL-to-Code
*/

#include "Reports.h"

std::string Reports::getId() {    
    return this->id;
}

void Reports::setId(std::string value) {
    this->id = value;
}

int Reports::getCreated() {    
    return this->created;
}

void Reports::setCreated(int value) {
    this->created = value;
}

int Reports::getModified() {    
    return this->modified;
}

void Reports::setModified(int value) {
    this->modified = value;
}

std::string Reports::getTitle() {    
    return this->title;
}

void Reports::setTitle(std::string value) {
    this->title = value;
}

std::string Reports::getDescription() {    
    return this->description;
}

void Reports::setDescription(std::string value) {
    this->description = value;
}

bool Reports::isDeleted() {    
    return this->deleted;
}

void Reports::setDeleted(bool value) {
    this->deleted = value;
}

