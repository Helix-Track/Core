/*
    Assets.cpp
    Generated with 'sql2code' 1.0.0-SNAPSHOT
    https://github.com/red-elf/SQL-to-Code
*/

#include "Assets.h"

std::string Assets::getId() {    
    return this->id;
}

void Assets::setId(std::string value) {
    this->id = value;
}

std::string Assets::getUrl() {    
    return this->url;
}

void Assets::setUrl(std::string value) {
    this->url = value;
}

std::string Assets::getDescription() {    
    return this->description;
}

void Assets::setDescription(std::string value) {
    this->description = value;
}

int Assets::getCreated() {    
    return this->created;
}

void Assets::setCreated(int value) {
    this->created = value;
}

int Assets::getModified() {    
    return this->modified;
}

void Assets::setModified(int value) {
    this->modified = value;
}

bool Assets::isDeleted() {    
    return this->deleted;
}

void Assets::setDeleted(bool value) {
    this->deleted = value;
}

