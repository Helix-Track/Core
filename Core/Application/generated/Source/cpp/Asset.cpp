/*
    Asset.cpp
    Generated with 'sql2code' 0.0.2
    https://github.com/red-elf/SQL-to-Code
*/

#include "Asset.h"

std::string Asset::getId() {    
    return this->id;
}

void Asset::setId(std::string &value) {
    this->id = value;
}

std::string Asset::getUrl() {    
    return this->url;
}

void Asset::setUrl(std::string &value) {
    this->url = value;
}

std::string Asset::getDescription() {    
    return this->description;
}

void Asset::setDescription(std::string &value) {
    this->description = value;
}

int Asset::getCreated() {    
    return this->created;
}

void Asset::setCreated(int &value) {
    this->created = value;
}

int Asset::getModified() {    
    return this->modified;
}

void Asset::setModified(int &value) {
    this->modified = value;
}

bool Asset::isDeleted() {    
    return this->deleted;
}

void Asset::setDeleted(bool &value) {
    this->deleted = value;
}

