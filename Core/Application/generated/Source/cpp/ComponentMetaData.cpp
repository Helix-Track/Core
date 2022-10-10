/*
    ComponentMetaData.cpp
    Generated with 'sql2code' 0.0.2
    https://github.com/red-elf/SQL-to-Code
*/

#include "ComponentMetaData.h"

std::string ComponentMetaData::getId() {    
    return this->id;
}

void ComponentMetaData::setId(std::string &value) {
    this->id = value;
}

std::string ComponentMetaData::getComponentId() {    
    return this->componentId;
}

void ComponentMetaData::setComponentId(std::string &value) {
    this->componentId = value;
}

std::string ComponentMetaData::getProperty() {    
    return this->property;
}

void ComponentMetaData::setProperty(std::string &value) {
    this->property = value;
}

std::string ComponentMetaData::getValue() {    
    return this->value;
}

void ComponentMetaData::setValue(std::string &value) {
    this->value = value;
}

int ComponentMetaData::getCreated() {    
    return this->created;
}

void ComponentMetaData::setCreated(int &value) {
    this->created = value;
}

int ComponentMetaData::getModified() {    
    return this->modified;
}

void ComponentMetaData::setModified(int &value) {
    this->modified = value;
}

bool ComponentMetaData::isDeleted() {    
    return this->deleted;
}

void ComponentMetaData::setDeleted(bool &value) {
    this->deleted = value;
}

