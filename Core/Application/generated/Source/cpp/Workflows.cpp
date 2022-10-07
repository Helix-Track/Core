/*
    Workflows.cpp
    Generated with 'sql2code' 0.0.1
    https://github.com/red-elf/SQL-to-Code
*/

#include "Workflows.h"

std::string Workflows::getId() {    
    return this->id;
}

void Workflows::setId(std::string value) {
    this->id = value;
}

std::string Workflows::getTitle() {    
    return this->title;
}

void Workflows::setTitle(std::string value) {
    this->title = value;
}

std::string Workflows::getDescription() {    
    return this->description;
}

void Workflows::setDescription(std::string value) {
    this->description = value;
}

int Workflows::getCreated() {    
    return this->created;
}

void Workflows::setCreated(int value) {
    this->created = value;
}

int Workflows::getModified() {    
    return this->modified;
}

void Workflows::setModified(int value) {
    this->modified = value;
}

bool Workflows::isDeleted() {    
    return this->deleted;
}

void Workflows::setDeleted(bool value) {
    this->deleted = value;
}

