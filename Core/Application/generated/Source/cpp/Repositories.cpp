/*
    Repositories.cpp
    Generated with 'sql2code' 1.0.0-SNAPSHOT
    https://github.com/red-elf/SQL-to-Code
*/

#include "Repositories.h"

std::string Repositories::getId() {    
    return this->id;
}

void Repositories::setId(std::string value) {
    this->id = value;
}

std::string Repositories::getRepository() {    
    return this->repository;
}

void Repositories::setRepository(std::string value) {
    this->repository = value;
}

std::string Repositories::getDescription() {    
    return this->description;
}

void Repositories::setDescription(std::string value) {
    this->description = value;
}

std::string Repositories::getRepositoryTypeId() {    
    return this->repositoryTypeId;
}

void Repositories::setRepositoryTypeId(std::string value) {
    this->repositoryTypeId = value;
}

int Repositories::getCreated() {    
    return this->created;
}

void Repositories::setCreated(int value) {
    this->created = value;
}

int Repositories::getModified() {    
    return this->modified;
}

void Repositories::setModified(int value) {
    this->modified = value;
}

bool Repositories::isDeleted() {    
    return this->deleted;
}

void Repositories::setDeleted(bool value) {
    this->deleted = value;
}

