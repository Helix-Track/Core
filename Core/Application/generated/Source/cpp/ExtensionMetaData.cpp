/*
    ExtensionMetaData.cpp
    Generated with 'sql2code' 0.0.2-SNAPSHOT
    https://github.com/red-elf/SQL-to-Code
*/

#include "ExtensionMetaData.h"

std::string ExtensionMetaData::getId() {    
    return this->id;
}

void ExtensionMetaData::setId(std::string &value) {
    this->id = value;
}

std::string ExtensionMetaData::getExtensionId() {    
    return this->extensionId;
}

void ExtensionMetaData::setExtensionId(std::string &value) {
    this->extensionId = value;
}

std::string ExtensionMetaData::getProperty() {    
    return this->property;
}

void ExtensionMetaData::setProperty(std::string &value) {
    this->property = value;
}

std::string ExtensionMetaData::getValue() {    
    return this->value;
}

void ExtensionMetaData::setValue(std::string &value) {
    this->value = value;
}

int ExtensionMetaData::getCreated() {    
    return this->created;
}

void ExtensionMetaData::setCreated(int &value) {
    this->created = value;
}

int ExtensionMetaData::getModified() {    
    return this->modified;
}

void ExtensionMetaData::setModified(int &value) {
    this->modified = value;
}

bool ExtensionMetaData::isDeleted() {    
    return this->deleted;
}

void ExtensionMetaData::setDeleted(bool &value) {
    this->deleted = value;
}

