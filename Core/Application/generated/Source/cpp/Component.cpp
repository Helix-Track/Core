/*
    Component.cpp
    Generated with 'sql2code' 0.0.1
    https://github.com/red-elf/SQL-to-Code
*/

#include "Component.h"

std::string Component::getId() {    
    return this->id;
}

void Component::setId(std::string value) {
    this->id = value;
}

std::string Component::getTitle() {    
    return this->title;
}

void Component::setTitle(std::string value) {
    this->title = value;
}

std::string Component::getDescription() {    
    return this->description;
}

void Component::setDescription(std::string value) {
    this->description = value;
}

int Component::getCreated() {    
    return this->created;
}

void Component::setCreated(int value) {
    this->created = value;
}

int Component::getModified() {    
    return this->modified;
}

void Component::setModified(int value) {
    this->modified = value;
}

bool Component::isDeleted() {    
    return this->deleted;
}

void Component::setDeleted(bool value) {
    this->deleted = value;
}

