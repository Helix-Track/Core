/*
    ComponentsMetaData.cpp
    Generated with 'sql2code' 0.0.1
    https://github.com/red-elf/SQL-to-Code
*/

#include "ComponentsMetaData.h"

std::string ComponentsMetaData::getId() {    
    return this->id;
}

void ComponentsMetaData::setId(std::string value) {
    this->id = value;
}

std::string ComponentsMetaData::getComponentId() {    
    return this->componentId;
}

void ComponentsMetaData::setComponentId(std::string value) {
    this->componentId = value;
}

std::string ComponentsMetaData::getProperty() {    
    return this->property;
}

void ComponentsMetaData::setProperty(std::string value) {
    this->property = value;
}

std::string ComponentsMetaData::getValue() {    
    return this->value;
}

void ComponentsMetaData::setValue(std::string value) {
    this->value = value;
}

int ComponentsMetaData::getCreated() {    
    return this->created;
}

void ComponentsMetaData::setCreated(int value) {
    this->created = value;
}

int ComponentsMetaData::getModified() {    
    return this->modified;
}

void ComponentsMetaData::setModified(int value) {
    this->modified = value;
}

bool ComponentsMetaData::isDeleted() {    
    return this->deleted;
}

void ComponentsMetaData::setDeleted(bool value) {
    this->deleted = value;
}

