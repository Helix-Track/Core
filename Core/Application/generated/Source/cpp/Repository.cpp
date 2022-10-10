/*
    Repository.cpp
    Generated with 'sql2code' 0.0.2-SNAPSHOT
    https://github.com/red-elf/SQL-to-Code
*/

#include "Repository.h"

std::string Repository::getId() {    
    return this->id;
}

void Repository::setId(std::string &value) {
    this->id = value;
}

std::string Repository::getRepository() {    
    return this->repository;
}

void Repository::setRepository(std::string &value) {
    this->repository = value;
}

std::string Repository::getDescription() {    
    return this->description;
}

void Repository::setDescription(std::string &value) {
    this->description = value;
}

std::string Repository::getRepositoryTypeId() {    
    return this->repositoryTypeId;
}

void Repository::setRepositoryTypeId(std::string &value) {
    this->repositoryTypeId = value;
}

int Repository::getCreated() {    
    return this->created;
}

void Repository::setCreated(int &value) {
    this->created = value;
}

int Repository::getModified() {    
    return this->modified;
}

void Repository::setModified(int &value) {
    this->modified = value;
}

bool Repository::isDeleted() {    
    return this->deleted;
}

void Repository::setDeleted(bool &value) {
    this->deleted = value;
}

