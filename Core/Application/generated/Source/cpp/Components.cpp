/*
    Components.cpp
    Generated with 'sql2code' 1.0.0-SNAPSHOT
    https://github.com/red-elf/SQL-to-Code
*/

#include "Components.h"

std::string Components::getId() {    
    return this->id;
}

void Components::setId(std::string value) {
    this->id = value;
}

std::string Components::getTitle() {    
    return this->title;
}

void Components::setTitle(std::string value) {
    this->title = value;
}

std::string Components::getDescription() {    
    return this->description;
}

void Components::setDescription(std::string value) {
    this->description = value;
}

int Components::getCreated() {    
    return this->created;
}

void Components::setCreated(int value) {
    this->created = value;
}

int Components::getModified() {    
    return this->modified;
}

void Components::setModified(int value) {
    this->modified = value;
}

bool Components::isDeleted() {    
    return this->deleted;
}

void Components::setDeleted(bool value) {
    this->deleted = value;
}

