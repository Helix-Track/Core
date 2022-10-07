/*
    RepositoryTypes.cpp
    Generated with 'sql2code' 0.0.1
    https://github.com/red-elf/SQL-to-Code
*/

#include "RepositoryTypes.h"

std::string RepositoryTypes::getId() {    
    return this->id;
}

void RepositoryTypes::setId(std::string value) {
    this->id = value;
}

std::string RepositoryTypes::getTitle() {    
    return this->title;
}

void RepositoryTypes::setTitle(std::string value) {
    this->title = value;
}

std::string RepositoryTypes::getDescription() {    
    return this->description;
}

void RepositoryTypes::setDescription(std::string value) {
    this->description = value;
}

int RepositoryTypes::getCreated() {    
    return this->created;
}

void RepositoryTypes::setCreated(int value) {
    this->created = value;
}

int RepositoryTypes::getModified() {    
    return this->modified;
}

void RepositoryTypes::setModified(int value) {
    this->modified = value;
}

bool RepositoryTypes::isDeleted() {    
    return this->deleted;
}

void RepositoryTypes::setDeleted(bool value) {
    this->deleted = value;
}

