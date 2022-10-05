/*
    Extensions.cpp
    Generated with 'sql2code' 1.0.0-SNAPSHOT
    https://github.com/red-elf/SQL-to-Code
*/

#include "Extensions.h"

std::string Extensions::getId() {    
    return this->id;
}

void Extensions::setId(std::string value) {
    this->id = value;
}

int Extensions::getCreated() {    
    return this->created;
}

void Extensions::setCreated(int value) {
    this->created = value;
}

int Extensions::getModified() {    
    return this->modified;
}

void Extensions::setModified(int value) {
    this->modified = value;
}

std::string Extensions::getTitle() {    
    return this->title;
}

void Extensions::setTitle(std::string value) {
    this->title = value;
}

std::string Extensions::getDescription() {    
    return this->description;
}

void Extensions::setDescription(std::string value) {
    this->description = value;
}

std::string Extensions::getExtensionKey() {    
    return this->extensionKey;
}

void Extensions::setExtensionKey(std::string value) {
    this->extensionKey = value;
}

bool Extensions::isEnabled() {    
    return this->enabled;
}

void Extensions::setEnabled(bool value) {
    this->enabled = value;
}

bool Extensions::isDeleted() {    
    return this->deleted;
}

void Extensions::setDeleted(bool value) {
    this->deleted = value;
}

