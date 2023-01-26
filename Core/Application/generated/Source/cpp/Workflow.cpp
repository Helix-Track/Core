/*
    Workflow.cpp
    Generated with 'sql2code' 0.0.3
    https://github.com/red-elf/SQL-to-Code
*/

#include "Workflow.h"

std::string Workflow::getId() {    
    return this->id;
}

void Workflow::setId(std::string &value) {
    this->id = value;
}

std::string Workflow::getTitle() {    
    return this->title;
}

void Workflow::setTitle(std::string &value) {
    this->title = value;
}

std::string Workflow::getDescription() {    
    return this->description;
}

void Workflow::setDescription(std::string &value) {
    this->description = value;
}

int Workflow::getCreated() {    
    return this->created;
}

void Workflow::setCreated(int &value) {
    this->created = value;
}

int Workflow::getModified() {    
    return this->modified;
}

void Workflow::setModified(int &value) {
    this->modified = value;
}

bool Workflow::isDeleted() {    
    return this->deleted;
}

void Workflow::setDeleted(bool &value) {
    this->deleted = value;
}

