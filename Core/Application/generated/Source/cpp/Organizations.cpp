/*
    Organizations.cpp
    Generated with 'sql2code' 0.0.1
    https://github.com/red-elf/SQL-to-Code
*/

#include "Organizations.h"

std::string Organizations::getId() {    
    return this->id;
}

void Organizations::setId(std::string value) {
    this->id = value;
}

std::string Organizations::getTitle() {    
    return this->title;
}

void Organizations::setTitle(std::string value) {
    this->title = value;
}

std::string Organizations::getDescription() {    
    return this->description;
}

void Organizations::setDescription(std::string value) {
    this->description = value;
}

int Organizations::getCreated() {    
    return this->created;
}

void Organizations::setCreated(int value) {
    this->created = value;
}

int Organizations::getModified() {    
    return this->modified;
}

void Organizations::setModified(int value) {
    this->modified = value;
}

bool Organizations::isDeleted() {    
    return this->deleted;
}

void Organizations::setDeleted(bool value) {
    this->deleted = value;
}

