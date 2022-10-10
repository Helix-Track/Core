/*
    RepositoryType.cpp
    Generated with 'sql2code' 0.0.2
    https://github.com/red-elf/SQL-to-Code
*/

#include "RepositoryType.h"

std::string RepositoryType::getId() {    
    return this->id;
}

void RepositoryType::setId(std::string &value) {
    this->id = value;
}

std::string RepositoryType::getTitle() {    
    return this->title;
}

void RepositoryType::setTitle(std::string &value) {
    this->title = value;
}

std::string RepositoryType::getDescription() {    
    return this->description;
}

void RepositoryType::setDescription(std::string &value) {
    this->description = value;
}

int RepositoryType::getCreated() {    
    return this->created;
}

void RepositoryType::setCreated(int &value) {
    this->created = value;
}

int RepositoryType::getModified() {    
    return this->modified;
}

void RepositoryType::setModified(int &value) {
    this->modified = value;
}

bool RepositoryType::isDeleted() {    
    return this->deleted;
}

void RepositoryType::setDeleted(bool &value) {
    this->deleted = value;
}

