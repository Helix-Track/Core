/*
    Extension.cpp
    Generated with 'sql2code' 0.0.3
    https://github.com/red-elf/SQL-to-Code
*/

#include "Extension.h"

std::string Extension::getId() {    
    return this->id;
}

void Extension::setId(std::string &value) {
    this->id = value;
}

int Extension::getCreated() {    
    return this->created;
}

void Extension::setCreated(int &value) {
    this->created = value;
}

int Extension::getModified() {    
    return this->modified;
}

void Extension::setModified(int &value) {
    this->modified = value;
}

std::string Extension::getTitle() {    
    return this->title;
}

void Extension::setTitle(std::string &value) {
    this->title = value;
}

std::string Extension::getDescription() {    
    return this->description;
}

void Extension::setDescription(std::string &value) {
    this->description = value;
}

std::string Extension::getExtensionKey() {    
    return this->extensionKey;
}

void Extension::setExtensionKey(std::string &value) {
    this->extensionKey = value;
}

bool Extension::isEnabled() {    
    return this->enabled;
}

void Extension::setEnabled(bool &value) {
    this->enabled = value;
}

bool Extension::isDeleted() {    
    return this->deleted;
}

void Extension::setDeleted(bool &value) {
    this->deleted = value;
}

